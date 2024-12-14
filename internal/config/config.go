package config

import (
	"fmt"
	"sync"
)

type UpState string

const (
	PrivateKeyFile = "id_rsa_gobun"
	PublicKeyFile  = "id_rsa_gobun.pub"
)

// VendorConfig 存储vendor的配置信息
type VendorConfig struct {
	APISecret map[string]string
}

// VendorConfigs 管理所有vendor的配置
type VendorConfigs struct {
	configs map[string]VendorConfig
	mu      sync.RWMutex
}

// NewVendorConfigs 创建一个新的VendorConfigs实例
func NewVendorConfigs() *VendorConfigs {
	return &VendorConfigs{
		configs: make(map[string]VendorConfig),
	}
}

// AddVendorConfig 添加或更新vendor配置
func (vc *VendorConfigs) AddVendorConfig(vendorID string, config VendorConfig) {
	vc.mu.Lock()
	defer vc.mu.Unlock()
	vc.configs[vendorID] = config
}

// GetVendorConfig 获取指定vendor的配置
func (vc *VendorConfigs) GetVendorConfig(vendorID string) (VendorConfig, error) {
	vc.mu.RLock()
	defer vc.mu.RUnlock()

	config, exists := vc.configs[vendorID]
	if !exists {
		return VendorConfig{}, fmt.Errorf("vendor config not found for ID: %s", vendorID)
	}
	return config, nil
}

// RemoveVendorConfig 删除指定vendor的配置
func (vc *VendorConfigs) RemoveVendorConfig(vendorID string) {
	vc.mu.Lock()
	defer vc.mu.Unlock()
	delete(vc.configs, vendorID)
}

// ListVendorConfigs 列出所有vendor的配置
func (vc *VendorConfigs) ListVendorConfigs() map[string]VendorConfig {
	vc.mu.RLock()
	defer vc.mu.RUnlock()

	// 创建配置的副本以避免并发访问问题
	configsCopy := make(map[string]VendorConfig)
	for k, v := range vc.configs {
		configsCopy[k] = v
	}
	return configsCopy
}
