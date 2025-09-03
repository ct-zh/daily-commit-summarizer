package config

import "context"

// ConfigLoader 配置加载器接口
type ConfigLoader interface {
	// Load 加载配置
	Load(ctx context.Context) (*Config, error)
	// Reload 重新加载配置
	Reload(ctx context.Context) (*Config, error)
	// Watch 监听配置变化
	Watch(ctx context.Context, callback func(*Config)) error
}

// ConfigValidator 配置验证器接口
type ConfigValidator interface {
	// Validate 验证配置
	Validate(config *Config) error
}

// ConfigProvider 配置提供者接口
type ConfigProvider interface {
	// GetConfig 获取当前配置
	GetConfig() *Config
	// IsLoaded 检查配置是否已加载
	IsLoaded() bool
}