package app

import (
	"fmt"
	"strings"

	gobun_config "github.com/funstory-ai/gobun/internal/config"
	"github.com/funstory-ai/gobun/vendors"
	"github.com/urfave/cli/v2"
)

var CommandVendor = &cli.Command{
	Name:  "vendor",
	Usage: "Manage vendor resources",
	Subcommands: []*cli.Command{
		{
			Name:  "add",
			Usage: "add a new vendor",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "provider",
					Usage:    "Specify the cloud provider",
					Required: true,
				},
				&cli.StringFlag{
					Name:  "values",
					Usage: "Specify key-value pairs (format: key1=value1,key2=value2)",
				},
				&cli.StringFlag{
					Name:  "name",
					Usage: "Specify the vendor name",
				},
			}, Action: addVendor,
		},
	},
}

func addVendor(ctx *cli.Context) error {
	provider := ctx.String("provider")
	valuesStr := ctx.String("values")
	name := ctx.String("name")

	// Validate provider type
	if provider != string(vendors.CloudProviderTypeXianGongYun) {
		return fmt.Errorf("invalid provider type: %s. Supported providers: %s",
			provider,
			vendors.CloudProviderTypeXianGongYun)
	}

	// Parse values string into map
	values := make(map[string]string)
	if valuesStr != "" {
		pairs := strings.Split(valuesStr, ",")
		for _, pair := range pairs {
			kv := strings.SplitN(pair, "=", 2)
			if len(kv) == 2 {
				values[kv[0]] = kv[1]
			}
		}
	}

	// If name is not provided, use default name
	if name == "" {
		name = "xiangongyun"
	}

	// 创建配置管理器
	vendorConfigs := gobun_config.NewVendorConfigs()

	// 创建 vendor 配置
	vendorConfig := gobun_config.VendorConfig{
		APISecret: values, // 使用解析后的 values 作为 APISecret
	}

	// 添加配置
	vendorConfigs.AddVendorConfig(name, vendorConfig)

	fmt.Printf("Successfully added vendor '%s' with provider '%s'\n", name, provider)
	return nil
}
