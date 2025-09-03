package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gopkg.in/yaml.v2"
)

func TestDefaultConfigLoader(t *testing.T) {
	// 设置测试环境变量
	os.Setenv("GO_ENV", "test")
	defer os.Unsetenv("GO_ENV")
	
	// 创建临时目录用于测试
	tempDir, err := os.MkdirTemp("", "config-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// 测试创建加载器
	t.Run("创建加载器", func(t *testing.T) {
		configPath := filepath.Join(tempDir, "config.yaml")
		loader := NewDefaultConfigLoader(configPath)
		assert.NotNil(t, loader)
		assert.False(t, loader.IsLoaded())
	})

	// 测试加载不存在的配置文件（应直接报错）
	t.Run("加载不存在的配置文件", func(t *testing.T) {
		configPath := filepath.Join(tempDir, "not-exist.yaml")
		loader := NewDefaultConfigLoader(configPath)

		// 加载不存在的配置文件应该报错
		cfg, err := loader.Load(context.Background())
		assert.Error(t, err)
		assert.Nil(t, cfg)
		assert.False(t, loader.IsLoaded())
		assert.Contains(t, err.Error(), "读取配置文件失败: not exist")

		// 验证文件不应该被创建
		_, err = os.Stat(configPath)
		assert.True(t, os.IsNotExist(err))
	})

	// 测试加载YAML配置文件
	t.Run("加载YAML配置文件", func(t *testing.T) {
		
		configPath := filepath.Join(tempDir, "config.yaml")
		testConfig := createTestConfig()

		// 创建YAML配置文件
		data, err := yaml.Marshal(testConfig)
		require.NoError(t, err)
		err = os.WriteFile(configPath, data, 0644)
		require.NoError(t, err)

		// 加载配置
		loader := NewDefaultConfigLoader(configPath)
		cfg, err := loader.Load(context.Background())
		assert.NoError(t, err)
		assert.NotNil(t, cfg)

		// 验证配置值
		assert.Equal(t, testConfig.Git.RepoPath, cfg.Git.RepoPath)
		assert.Equal(t, testConfig.LLM.Model, cfg.LLM.Model)
		assert.Equal(t, testConfig.Notification.Lark.WebhookURL, cfg.Notification.Lark.WebhookURL)
		assert.Equal(t, testConfig.App.Name, cfg.App.Name)
		assert.Equal(t, testConfig.Logger.Level, cfg.Logger.Level)
	})

	// 测试重新加载配置
	t.Run("重新加载配置", func(t *testing.T) {
		
		configPath := filepath.Join(tempDir, "reload.yaml")
		testConfig := createTestConfig()

		// 创建初始配置文件
		data, err := yaml.Marshal(testConfig)
		require.NoError(t, err)
		err = os.WriteFile(configPath, data, 0644)
		require.NoError(t, err)

		// 加载配置
		loader := NewDefaultConfigLoader(configPath)
		cfg1, err := loader.Load(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, "test-app", cfg1.App.Name)

		// 修改配置文件
		testConfig.App.Name = "updated-app"
		data, err = yaml.Marshal(testConfig)
		require.NoError(t, err)
		err = os.WriteFile(configPath, data, 0644)
		require.NoError(t, err)

		// 重新加载配置
		cfg2, err := loader.Reload(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, "updated-app", cfg2.App.Name)
	})

	// 测试GetConfig方法
	t.Run("GetConfig方法", func(t *testing.T) {
		configPath := filepath.Join(tempDir, "get-config.yaml")
		testConfig := createTestConfig()

		// 创建配置文件
		data, err := yaml.Marshal(testConfig)
		require.NoError(t, err)
		err = os.WriteFile(configPath, data, 0644)
		require.NoError(t, err)

		// 加载配置
		loader := NewDefaultConfigLoader(configPath)
		_, err = loader.Load(context.Background())
		assert.NoError(t, err)

		// 获取配置
		cfg := loader.GetConfig()
		assert.NotNil(t, cfg)
		assert.Equal(t, testConfig.App.Name, cfg.App.Name)
	})
}

// 创建测试配置
func createTestConfig() *Config {
	return &Config{
		Git: GitConfig{
			RepoPath:        "./test-repo",
			RemoteName:      "origin",
			DefaultBranch:   "main",
			ExcludePatterns: []string{"*.lock", "*.log"},
			MaxCommits:      100,
		},
		LLM: LLMConfig{
			APIKey:      "test-api-key",
			BaseURL:     "https://api.test.com",
			Model:       "test-model",
			MaxTokens:   1000,
			Temperature: 0.5,
			Timeout:     10 * time.Second,
			RetryCount:  3,
		},
		Notification: NotificationConfig{
			Lark: LarkConfig{
				WebhookURL: "https://test-webhook.com",
				Secret:     "test-secret",
				Timeout:    5 * time.Second,
			},
			Enabled: true,
		},
		App: AppConfig{
			Name:        "test-app",
			Version:     "1.0.0",
			Environment: "testing",
			Debug:       true,
			WorkDir:     "./test-work",
			DataDir:     "./test-data",
		},
		Logger: LoggerConfig{
			Level:      "debug",
			Format:     "json",
			Output:     "stdout",
			FilePath:   "./test-logs/app.log",
			MaxSize:    10,
			MaxBackups: 5,
			MaxAge:     7,
			Compress:   true,
		},
	}
}