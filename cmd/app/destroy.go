package app

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	gobun_config "github.com/funstory-ai/gobun/internal/config"
	"github.com/funstory-ai/gobun/internal/resource"
	"github.com/funstory-ai/gobun/vendors"
	"github.com/urfave/cli/v2"
)

var CommandDestroy = &cli.Command{
	Name:      "destroy",
	Usage:     "销毁一个或多个 pods",
	ArgsUsage: "<pod-id> [pod-id ...]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "vendor",
			Usage:    "Specify the vendor",
			Required: true,
		},
	},
	Action: destroy,
}

func destroy(ctx *cli.Context) error {
	// 检查是否提供了至少一个 pod ID
	if ctx.NArg() < 1 {
		return fmt.Errorf("至少需要一个 pod ID")
	}

	// 获取 vendor 配置
	vendorId := ctx.String("vendor")
	vendorConfigs, err := gobun_config.NewVendorConfigs()
	if err != nil {
		return fmt.Errorf("failed to create vendor configs: %w", err)
	}
	vendorConfig, err := vendorConfigs.GetVendorConfig(vendorId)
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

		// 创建 pod 实例
		pod := resource.Pod{
			ID: podID,
		}

		// 启动进度显示
		done := make(chan bool)
		go func() {
			for {
				select {
				case <-done:
					return
				default:
					fmt.Print(".")
					time.Sleep(500 * time.Millisecond)
				}
			}
		}()

		// 执行销毁
		err = vendor.DestroyPod(ctx.Context, pod)
		done <- true // 停止进度显示

		if err != nil {
			fmt.Printf("\n销毁 pod %s 失败: %v\n", podID, err)
		} else {
			fmt.Printf("\n成功销毁 pod: %s\n", podID)
		}
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
