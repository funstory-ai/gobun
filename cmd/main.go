package main

import (
	"os"

	"github.com/funstory-ai/gobun/internal/ssh"
)

func main() {
	sshClient, err := ssh.NewClient(ssh.Options{
		User:           "root",
		Server:         "8.146.199.203",
		Port:           22,
		PrivateKeyPath: os.Getenv("HOME") + "/.ssh/id_rsa",
		Auth:           true,
	})
	if err != nil {
		panic(err)
	}
	sshClient.Attach()
}
