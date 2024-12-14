package app

import (
	"fmt"
	"strconv"
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

	// 创建配置管理器
	vendorConfigs, err := gobun_config.NewVendorConfigs()
	if err != nil {
		return fmt.Errorf("failed to create vendor configs: %w", err)
	}

	// If name is not provided, generate a default name
	if name == "" {
		// 找到现有同类型vendor中最大的序号
		maxNum := 0
		for existingName := range vendorConfigs.ListVendorConfigs() {
			if strings.HasPrefix(existingName, provider+"-") {
				if num, err := strconv.Atoi(strings.TrimPrefix(existingName, provider+"-")); err == nil {
					if num > maxNum {
						maxNum = num
					}
				}
			}
		}
		name = fmt.Sprintf("%s-%d", provider, maxNum+1)
	}

	// 创建 vendor 配置
	vendorConfig := gobun_config.VendorConfig{
		Provider:  provider,
		APISecret: values,
	}

	// 添加配置
	vendorConfigs.AddVendorConfig(name, vendorConfig)

	fmt.Printf("Successfully added vendor '%s' with provider '%s'\n", name, provider)
	return nil
}
