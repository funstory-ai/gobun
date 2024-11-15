package app

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/funstory-ai/gobun/adaptors/xiangongyun"
	"github.com/funstory-ai/gobun/internal"
	"github.com/funstory-ai/gobun/internal/ssh"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var CommandUp = &cli.Command{
	Name:   "up",
	Usage:  "Quickly start a pod and attach to it",
	Action: up,
}

func up(ctx *cli.Context) error {
	// Get token from environment
	token := os.Getenv("XGY_TOKEN")
	if token == "" {
		return fmt.Errorf("环境变量 XGY_TOKEN 未设置，请设置后重试")
	}

	pool := xiangongyun.NewPool("Bearer " + token)

	// Create pod with default options
	options := internal.PodOptions{
		GPUModel: internal.GPUModelRTX4090,
		GPUCount: 1,
	}

	fmt.Println("Creating pod...")
	pod, err := pool.CreatePod(options)
	if err != nil {
		return fmt.Errorf("failed to create pod: %w", err)
	}

	// Set up signal handling for cleanup
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start goroutine to handle cleanup on signal
	go func() {
		<-sigChan
		fmt.Println("\nReceived signal, cleaning up...")
		if err := pool.DestroyPod(pod.ID); err != nil {
			logrus.Errorf("Failed to destroy pod: %v", err)
		}
		os.Exit(0)
	}()

	// Defer pod cleanup in case of any errors
	defer func() {
		fmt.Println("Cleaning up pod...")
		if err := pool.DestroyPod(pod.ID); err != nil {
			logrus.Errorf("Failed to destroy pod: %v", err)
		}
	}()

	fmt.Printf("Pod created successfully (ID: %s)\n", pod.ID)
	fmt.Println("Waiting for pod to be ready...")

	// Poll pod status until it's running
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		pod, err = pool.GetPod(pod.ID)
		if err != nil {
			return fmt.Errorf("failed to get pod status: %w", err)
		}

		if pod.Status == string(internal.StatusRunning) {
			fmt.Println("Pod is now running!")
			break
		} else if pod.Status == string(internal.StatusError) {
			return fmt.Errorf("pod failed to start")
		}

		fmt.Printf("Current status: %s\n", pod.Status)
	}

	fmt.Println("Attaching to pod...")
	// Create SSH client options
	port, err := strconv.Atoi(pod.SSHPort)
	if err != nil {
		return fmt.Errorf("failed to parse SSH port: %w", err)
	}
	opt := ssh.Options{
		Server:   pod.SSHDomain,
		Port:     port,
		User:     pod.SSHUser,
		Password: pod.Password,
		Auth:     true,
	}

	// Create new SSH client
	client, err := ssh.NewClient(opt)
	if err != nil {
		return fmt.Errorf("failed to create SSH client: %w", err)
	}
	defer client.Close()

	// Attach to the pod
	if err := client.Attach(); err != nil {
		return fmt.Errorf("failed to attach to pod: %w", err)
	}

	return nil
}
