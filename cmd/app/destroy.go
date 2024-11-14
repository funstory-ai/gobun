package app

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/funstory-ai/gobun/adaptors/xiangongyun"
	"github.com/urfave/cli/v2"
)

const (
	// EnvXGYToken 用于存储环境变量名称
	EnvXGYToken = "XGY_TOKEN"
)

// CommandDestroy 定义了 destroy 命令
var CommandDestroy = &cli.Command{
	Name:      "destroy",
	Usage:     "销毁一个或多个 pods",
	ArgsUsage: "<pod-id> [pod-id ...]",
	Action:    destroy,
}

func destroy(ctx *cli.Context) error {
	// 获取 XGY_TOKEN 环境变量
	token := os.Getenv(EnvXGYToken)
	if token == "" {
		return fmt.Errorf("环境变量 %s 未设置，请设置后重试", EnvXGYToken)
	}

	pool := xiangongyun.NewPool("Bearer " + token)

	// 检查是否提供了至少一个 pod ID
	if ctx.NArg() < 1 {
		return fmt.Errorf("至少需要一个 pod ID")
	}

	// 遍历所有提供的 pod ID 并尝试销毁
	for _, podID := range ctx.Args().Slice() {
		fmt.Printf("准备销毁 pod: %s\n", podID)

		// 交互确认
		confirm, err := getConfirmation(fmt.Sprintf("您确定要销毁 pod %s 吗？(y/N): ", podID))
		if err != nil {
			fmt.Printf("获取确认失败: %v\n", err)
			continue
		}
		if !confirm {
			fmt.Printf("跳过销毁 pod: %s\n", podID)
			continue
		}

		// 显示销毁进度
		fmt.Printf("正在销毁 pod: %s...\n", podID)
		statusCh := make(chan string)
		go func(podID string, statusCh chan<- string) {
			err := pool.DestroyPod(podID)
			if err != nil {
				statusCh <- fmt.Sprintf("销毁 pod %s 失败: %v", podID, err)
				return
			}
			statusCh <- fmt.Sprintf("成功销毁 pod: %s", podID)
		}(podID, statusCh)

		// 模拟进度条
		go showProgress()

		// 等待销毁结果
		result := <-statusCh
		fmt.Println("\n" + result)
	}

	return nil
}

// getConfirmation 提示用户确认操作
func getConfirmation(message string) (bool, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(message)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes", nil
}

// showProgress 显示简单的进度指示器
func showProgress() {
	for i := 0; i < 5; i++ {
		fmt.Print(".")
		time.Sleep(500 * time.Millisecond)
	}
}
