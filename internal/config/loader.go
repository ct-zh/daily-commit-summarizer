package config

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v2"
)

// DefaultConfigLoader 默认配置加载器
type DefaultConfigLoader struct {
	mu         sync.RWMutex
	config     *Config
	configPath string
	loaded     bool
}

// NewDefaultConfigLoader 创建默认配置加载器
func NewDefaultConfigLoader(configPath string) *DefaultConfigLoader {
	return &DefaultConfigLoader{
		configPath: configPath,
	}
}

// Load 加载配置
func (l *DefaultConfigLoader) Load(ctx context.Context) (*Config, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// 如果配置文件不存在，直接报错
	if _, err := os.Stat(l.configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("读取配置文件失败: not exist")
	}

	// 读取配置文件
	data, err := ioutil.ReadFile(l.configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析配置
	config := &Config{}
	if err := l.parseConfig(data, config); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	// 验证配置
	validator := &DefaultConfigValidator{}
	if err := validator.Validate(config); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	l.config = config
	l.loaded = true

	return config, nil
}

// Reload 重新加载配置
func (l *DefaultConfigLoader) Reload(ctx context.Context) (*Config, error) {
	l.mu.Lock()
	l.loaded = false
	l.mu.Unlock()

	return l.Load(ctx)
}

// Watch 监听配置变化
func (l *DefaultConfigLoader) Watch(ctx context.Context, callback func(*Config)) error {
	// 简单的轮询实现，生产环境可以使用 fsnotify
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	lastModTime := time.Time{}
	if stat, err := os.Stat(l.configPath); err == nil {
		lastModTime = stat.ModTime()
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if stat, err := os.Stat(l.configPath); err == nil {
				if stat.ModTime().After(lastModTime) {
					lastModTime = stat.ModTime()
					if config, err := l.Reload(ctx); err == nil {
						callback(config)
					}
				}
			}
		}
	}
}

// GetConfig 获取当前配置
func (l *DefaultConfigLoader) GetConfig() *Config {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.config
}

// IsLoaded 检查配置是否已加载
func (l *DefaultConfigLoader) IsLoaded() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.loaded
}

// parseConfig 解析配置文件，只支持YAML格式
func (l *DefaultConfigLoader) parseConfig(data []byte, config *Config) error {
	ext := strings.ToLower(filepath.Ext(l.configPath))
	if ext != ".yaml" && ext != ".yml" {
		return fmt.Errorf("不支持的配置文件格式: %s，仅支持YAML格式", ext)
	}
	return yaml.Unmarshal(data, config)
}