package app

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	gobun_config "github.com/funstory-ai/gobun/internal/config"
	"github.com/funstory-ai/gobun/internal/resource"
	"github.com/funstory-ai/gobun/internal/ssh"
	"github.com/funstory-ai/gobun/vendors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var CommandUp = &cli.Command{
	Name:   "up",
	Usage:  "Quickly start a pod and attach to it",
	Action: up,
}

func up(ctx *cli.Context) error {
	// 获取 vendor 配置
	vendorConfig, err := gobun_config.NewVendorConfigs().GetVendorConfig("xiangongyun") // 或从命令行参数获取
	if err != nil {
		return fmt.Errorf("failed to get vendor config: %w", err)
	}

	// 创建 vendor 实例
	options := vendors.VendorOptions{
		APISecret: vendorConfig.APISecret,
	}
	vendor, err := vendors.NewVendor(ctx.Context, vendors.CloudProviderTypeXianGongYun, options)
	if err != nil {
		return fmt.Errorf("failed to create vendor: %w", err)
	}

	// 创建 pod
	podOptions := resource.PodOptions{
		GPUModel: resource.GPUModelRTX4090,
		GPUCount: 1,
	}
	pod, err := vendor.CreatePod(ctx.Context, podOptions)
	if err != nil {
		return fmt.Errorf("failed to create pod: %w", err)
	}

	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 处理清理的 goroutine
	go func() {
		<-sigChan
		fmt.Println("\nReceived signal, cleaning up...")
		if err := vendor.DestroyPod(ctx.Context, pod); err != nil {
			logrus.Errorf("Failed to destroy pod: %v", err)
		}
		os.Exit(0)
	}()

	// 延迟清理 pod
	defer func() {
		fmt.Println("Cleaning up pod...")
		if err := vendor.DestroyPod(ctx.Context, pod); err != nil {
			logrus.Errorf("Failed to destroy pod: %v", err)
		}
	}()

	fmt.Printf("Pod created successfully (ID: %s)\n", pod.ID)
	fmt.Println("Waiting for pod to be ready...")

	// 轮询 pod 状态
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		pod, err = vendor.GetPod(ctx.Context, pod)
		if err != nil {
			return fmt.Errorf("failed to get pod status: %w", err)
		}

		if pod.Status == string(resource.StatusRunning) {
			fmt.Println("Pod is now running!")
			break
		} else if pod.Status == string(resource.StatusError) {
			return fmt.Errorf("pod failed to start")
		}

		fmt.Printf("Current status: %s\n", pod.Status)
	}

	fmt.Println("Attaching to pod...")
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

	client, err := ssh.NewClient(opt)
	if err != nil {
		return fmt.Errorf("failed to create SSH client: %w", err)
	}
	defer client.Close()

	if err := client.Attach(); err != nil {
		return fmt.Errorf("failed to attach to pod: %w", err)
	}

	return nil
}
