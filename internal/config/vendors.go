package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

const (
	PrivateKeyFile = "id_rsa_gobun"
	PublicKeyFile  = "id_rsa_gobun.pub"
	configFileName = "vendors.yaml"
)

// VendorConfig 存储vendor的配置信息
type VendorConfig struct {
	Provider  string            `yaml:"provider"`
	APISecret map[string]string `yaml:"api_secret"`
}

// VendorConfigs 管理所有vendor的配置
type VendorConfigs struct {
	configs    map[string]VendorConfig `yaml:"configs"`
	mu         sync.RWMutex
	configPath string
}

// getConfigDir 获取XDG配置目录
func getConfigDir() (string, error) {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configHome = filepath.Join(homeDir, ".config")
	}
	configDir := filepath.Join(configHome, "gobun")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}
	return configDir, nil
}

// NewVendorConfigs 创建一个新的VendorConfigs实例
func NewVendorConfigs() (*VendorConfigs, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return nil, err
	}

	vc := &VendorConfigs{
		configs:    make(map[string]VendorConfig),
		configPath: filepath.Join(configDir, configFileName),
	}

	// 如果配置文件存在，则加载它
	if _, err := os.Stat(vc.configPath); err == nil {
		if err := vc.load(); err != nil {
			return nil, err
		}
	}

	return vc, nil
}

// load 从文件加载配置
func (vc *VendorConfigs) load() error {
	data, err := os.ReadFile(vc.configPath)
	if err != nil {
		return err
	}

	vc.mu.Lock()
	defer vc.mu.Unlock()

	return yaml.Unmarshal(data, &vc.configs)
}

// save 保存配置到文件
func (vc *VendorConfigs) save() error {
	vc.mu.RLock()
	defer vc.mu.RUnlock()

	data, err := yaml.Marshal(vc.configs)
	if err != nil {
		return err
	}

	return os.WriteFile(vc.configPath, data, 0600)
}

// AddVendorConfig 添加或更新vendor配置
func (vc *VendorConfigs) AddVendorConfig(vendorID string, config VendorConfig) error {
	vc.mu.Lock()
	vc.configs[vendorID] = config
	vc.mu.Unlock()
	return vc.save()
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
func (vc *VendorConfigs) RemoveVendorConfig(vendorID string) error {
	vc.mu.Lock()
	delete(vc.configs, vendorID)
	vc.mu.Unlock()
	return vc.save()
}

// ListVendorConfigs 列出所有vendor的配置
func (vc *VendorConfigs) ListVendorConfigs() map[string]VendorConfig {
	vc.mu.RLock()
	defer vc.mu.RUnlock()

	configsCopy := make(map[string]VendorConfig)
	for k, v := range vc.configs {
		configsCopy[k] = v
	}
	return configsCopy
}
