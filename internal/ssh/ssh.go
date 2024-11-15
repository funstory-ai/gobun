package ssh

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/cockroachdb/errors"
	"github.com/funstory-ai/gobun/internal/ssh/config"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/term"
)

type Client interface {
	Attach() error
	ExecWithOutput(cmd string) ([]byte, error)
	LocalForward(localAddress, targetAddress string) error
	RemoteForward(localAddress, targetAddress string) error
	Close() error
}

type Options struct {
	AgentForwarding bool
	Server          string
	User            string
	Port            int
	Auth            bool
	PrivateKeyPath  string
	PrivateKeyPwd   string
	Password        string
}

func DefaultOptions() Options {
	return Options{
		User:            "envd",
		Auth:            true,
		PrivateKeyPwd:   "",
		AgentForwarding: true,
	}
}

func GetOptions(entry string) (*Options, error) {
	path, err := config.GetPrivateKey()
	if err != nil {
		return nil, errors.Wrap(err, "getting private key failed")
	}
	port := 22
	opt := DefaultOptions()
	opt.Port = port
	opt.PrivateKeyPath = path
	return &opt, nil
}

type generalClient struct {
	cli *ssh.Client
	opt *Options
}

func NewClient(opt Options) (Client, error) {
	logger := logrus.WithFields(logrus.Fields{
		"user":             opt.User,
		"port":             opt.Port,
		"server":           opt.Server,
		"agent-forwarding": opt.AgentForwarding,
		"auth":             opt.Auth,
	})
	logger.Debug("ssh to the environment")

	config := &ssh.ClientConfig{
		User: opt.User,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			// use OpenSSH's known_hosts file if you care about host validation
			return nil
		},
	}

	var cli *ssh.Client

	if opt.Auth {
		if opt.Password != "" {
			// Use password authentication if provided
			config.Auth = []ssh.AuthMethod{
				ssh.Password(opt.Password),
			}
		} else {
			// Existing private key authentication
			pemBytes, err := os.ReadFile(opt.PrivateKeyPath)
			if err != nil {
				return nil, errors.Wrapf(
					err, "reading private key %s failed", opt.PrivateKeyPath)
			}
			signer, err := signerFromPem(pemBytes, []byte(opt.PrivateKeyPwd))
			if err != nil {
				return nil, errors.Wrap(err, "creating signer from private key failed")
			}
			config.Auth = []ssh.AuthMethod{
				ssh.PublicKeys(signer),
			}
		}
	}

	host := fmt.Sprintf("%s:%d", opt.Server, opt.Port)
	// open connection
	conn, err := ssh.Dial("tcp", host, config)
	if err != nil {
		return nil, errors.Wrap(err, "dialing failed")
	}
	cli = conn

	if opt.AgentForwarding {
		// open connection to the local agent
		socketLocation := os.Getenv("SSH_AUTH_SOCK")
		if socketLocation != "" {
			agentConn, err := net.Dial("unix", socketLocation)
			if err != nil {
				return nil, errors.Wrap(err, "could not connect to local agent socket")
			}
			// create agent and add in auth
			forwardingAgent := agent.NewClient(agentConn)
			// add callback for forwarding agent to SSH config
			// might want to handle reconnects appending multiple callbacks
			auth := ssh.PublicKeysCallback(forwardingAgent.Signers)
			config.Auth = append(config.Auth, auth)
			if err := agent.ForwardToAgent(cli, forwardingAgent); err != nil {
				return nil, errors.Wrap(err, "forwarding agent to client failed")
			}
		} else {
			logger.Warn("SSH Agent Forwarding is disabled. This will have no impact on your normal use if you do not use the ssh key on the host.")
		}
	}

	return &generalClient{
		cli: cli,
		opt: &opt,
	}, nil
}

func (c generalClient) Close() error {
	return c.cli.Close()
}

func (c generalClient) ExecWithOutput(cmd string) ([]byte, error) {
	defer c.cli.Close()

	// open session
	session, err := c.cli.NewSession()
	if err != nil {
		return nil, errors.Wrap(err, "creating session failed")
	}
	defer session.Close()

	if c.opt.AgentForwarding {
		if err := agent.RequestAgentForwarding(session); err != nil {
			return nil, errors.Wrap(err, "requesting agent forwarding failed")
		}
	}

	return session.CombinedOutput(cmd)
}

func (c generalClient) Attach() error {
	// open session
	session, err := c.cli.NewSession()
	if err != nil {
		return errors.Wrap(err, "creating session failed")
	}
	defer session.Close()

	if c.opt.AgentForwarding {
		if err := agent.RequestAgentForwarding(session); err != nil {
			return errors.Wrap(err, "requesting agent forwarding failed")
		}
	}

	logger := logrus.WithFields(logrus.Fields{
		"user":             c.opt.User,
		"port":             c.opt.Port,
		"server":           c.opt.Server,
		"agent-forwarding": c.opt.AgentForwarding,
		"auth":             c.opt.Auth,
	})

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,     // Enable echoing (changed from 0)
		ssh.ECHOCTL:       1,     // Print control chars (changed from 0)
		ssh.IGNCR:         0,     // Don't ignore CR on input (changed from 1)
		ssh.TTY_OP_ISPEED: 14400, // Input speed in baud
		ssh.TTY_OP_OSPEED: 14400, // Output speed in baud
	}

	height, width := 80, 40
	var termFD int
	var ok bool
	if termFD, ok = isTerminal(os.Stdin); ok {
		width, height, err = term.GetSize(int(os.Stdout.Fd()))
		logger.Debugf("terminal width %d height %d", width, height)
		if err != nil {
			logger.WithError(err).Debug("request for terminal size failed")
		}
	}

	state, err := term.MakeRaw(termFD)
	if err != nil {
		logger.WithError(err).Debug("request for raw terminal failed")
	}

	defer func() {
		if state == nil {
			return
		}

		if err := term.Restore(termFD, state); err != nil {
			logger.WithError(err).Debugf("failed to restore terminal")
		}

		logger.Debugf("terminal restored")
	}()

	if err := session.RequestPty("xterm-256color", height, width, modes); err != nil {
		return errors.Newf("request for pseudo terminal failed: %w", err)
	}

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	logger.Debug("starting shell")
	err = session.Shell()
	if err != nil {
		return errors.Wrap(err, "starting shell failed")
	}
	logger.Debug("waiting for shell to exit")
	if err = session.Wait(); err != nil {
		var ee *ssh.ExitError
		if ok := errors.As(err, &ee); ok {
			switch ee.ExitStatus() {
			case 130:
				return nil
			case 137:
				logger.Warn(`Insufficient memory.`)
			}
		}
		var emr *ssh.ExitMissingError
		if ok := errors.As(err, &emr); ok {
			logger.WithError(emr).Debug("exit status missing")
			return nil
		}
		return errors.Wrap(err, "waiting for session failed")
	}

	logger.Debug("shell exited")
	return nil
}

func (c generalClient) LocalForward(localAddress, targetAddress string) error {
	localListener, err := net.Listen("tcp", localAddress)
	if err != nil {
		return errors.Wrap(err, "net.Listen failed")
	}

	logger := logrus.WithField("type", "local")

	logger.Debugf("begin to forward %s to %s", localAddress, targetAddress)
	for {
		localCon, err := localListener.Accept()
		if err != nil {
			return errors.Wrap(err, "listen.Accept failed")
		}

		sshConn, err := c.cli.Dial("tcp", targetAddress)
		if err != nil {
			return errors.Wrap(err, "listen.Accept failed")
		}

		// Copy local.Reader to sshConn.Writer
		go func() {
			_, err = io.Copy(sshConn, localCon)
			if err != nil {
				logger.WithError(err).Debug("io.Copy failed")
			}
		}()

		// Copy sshConn.Reader to localCon.Writer
		go func() {
			_, err = io.Copy(localCon, sshConn)
			if err != nil {
				logger.WithError(err).Debug("io.Copy failed")
			}
		}()
	}
}

func (c generalClient) RemoteForward(remoteAddress, targetAddress string) error {
	sshListener, err := c.cli.Listen("tcp", remoteAddress)
	if err != nil {
		return errors.Wrap(err, "cli.Listen failed")
	}

	logger := logrus.WithField("type", "remote")

	logger.Debugf("begin to forward %s to %s", remoteAddress, targetAddress)
	for {
		sshCon, err := sshListener.Accept()
		if err != nil {
			return errors.Wrap(err, "listen.Accept failed")
		}

		targetCon, err := net.Dial("tcp", targetAddress)
		if err != nil {
			return errors.Wrap(err, "net.Dial failed")
		}

		// Copy sshCon.Reader to targetCon.Writer
		go func() {
			_, err = io.Copy(targetCon, sshCon)
			if err != nil {
				logger.WithError(err).Debug("io.Copy failed")
			}
		}()

		// Copy targetCon.Reader to sshCon.Writer
		go func() {
			_, err = io.Copy(sshCon, targetCon)
			if err != nil {
				logger.WithError(err).Debug("io.Copy failed")
			}
		}()
	}
}

func isTerminal(r io.Reader) (int, bool) {
	switch v := r.(type) {
	case *os.File:
		return int(v.Fd()), term.IsTerminal(int(v.Fd()))
	default:
		return 0, false
	}
}

func signerFromPem(pemBytes []byte, password []byte) (ssh.Signer, error) {
	// read pem block
	err := errors.New("Pem decode failed, no key found")
	pemBlock, _ := pem.Decode(pemBytes)
	if pemBlock == nil {
		return nil, err
	}

	// handle encrypted key
	// nolint
	if x509.IsEncryptedPEMBlock(pemBlock) {
		// decrypt PEM
		// nolint
		pemBlock.Bytes, err = x509.DecryptPEMBlock(pemBlock, []byte(password))
		if err != nil {
			return nil, errors.Newf("decrypting PEM block failed %w", err)
		}

		// get RSA, EC or DSA key
		key, err := parsePemBlock(pemBlock)
		if err != nil {
			return nil, err
		}

		// generate signer instance from key
		signer, err := ssh.NewSignerFromKey(key)
		if err != nil {
			return nil, errors.Newf("creating signer from encrypted key failed %w", err)
		}

		return signer, nil
	} else {
		// generate signer instance from plain key
		signer, err := ssh.ParsePrivateKey(pemBytes)
		if err != nil {
			return nil, errors.Newf("parsing plain private key failed %w", err)
		}

		return signer, nil
	}
}

func parsePemBlock(block *pem.Block) (interface{}, error) {
	switch block.Type {
	case "RSA PRIVATE KEY":
		key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, errors.Newf("Parsing PKCS private key failed %w", err)
		}
		return key, nil
	case "EC PRIVATE KEY":
		key, err := x509.ParseECPrivateKey(block.Bytes)
		if err != nil {
			return nil, errors.Newf("Parsing EC private key failed %w", err)
		}
		return key, nil
	case "DSA PRIVATE KEY":
		key, err := ssh.ParseDSAPrivateKey(block.Bytes)
		if err != nil {
			return nil, errors.Newf("Parsing DSA private key failed %w", err)
		}
		return key, nil
	default:
		return nil, errors.Newf("Parsing private key failed, unsupported key type %q", block.Type)
	}
}
