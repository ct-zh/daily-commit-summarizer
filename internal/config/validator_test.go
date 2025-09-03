package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfigValidator(t *testing.T) {
	validator := &DefaultConfigValidator{}

	// 测试空配置
	t.Run("空配置验证", func(t *testing.T) {
		err := validator.Validate(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "配置不能为空")
	})

	// 测试有效配置
	t.Run("有效配置验证", func(t *testing.T) {
		cfg := createValidConfig()
		err := validator.Validate(cfg)
		assert.NoError(t, err)
	})

	// 测试Git配置验证
	t.Run("Git配置验证", func(t *testing.T) {
		// 测试仓库路径为空
		cfg := createValidConfig()
		cfg.Git.RepoPath = ""
		err := validator.Validate(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "仓库路径不能为空")

		// 测试默认值设置
		cfg = createValidConfig()
		cfg.Git.RemoteName = ""
		cfg.Git.DefaultBranch = ""
		cfg.Git.MaxCommits = 0
		err = validator.Validate(cfg)
		assert.NoError(t, err)
		assert.Equal(t, "origin", cfg.Git.RemoteName)
		assert.Equal(t, "main", cfg.Git.DefaultBranch)
		assert.Equal(t, 100, cfg.Git.MaxCommits)
	})

	// 测试LLM配置验证
	t.Run("LLM配置验证", func(t *testing.T) {
		// 测试API密钥为空
		cfg := createValidConfig()
		cfg.LLM.APIKey = ""
		err := validator.Validate(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "LLM API密钥不能为空")

		// 测试温度参数范围
		cfg = createValidConfig()
		cfg.LLM.Temperature = 3.0
		err = validator.Validate(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "温度参数必须在0-2之间")

		// 测试默认值设置
		cfg = createValidConfig()
		cfg.LLM.BaseURL = ""
		cfg.LLM.Model = ""
		cfg.LLM.MaxTokens = 0
		cfg.LLM.Timeout = 0
		cfg.LLM.RetryCount = -1
		err = validator.Validate(cfg)
		assert.NoError(t, err)
		assert.Equal(t, "https://api.openai.com/v1", cfg.LLM.BaseURL)
		assert.Equal(t, "gpt-3.5-turbo", cfg.LLM.Model)
		assert.Equal(t, 4000, cfg.LLM.MaxTokens)
		assert.Equal(t, 30*time.Second, cfg.LLM.Timeout)
		assert.Equal(t, 3, cfg.LLM.RetryCount)
	})

	// 测试通知配置验证
	t.Run("通知配置验证", func(t *testing.T) {
		// 测试禁用通知
		cfg := createValidConfig()
		cfg.Notification.Enabled = false
		cfg.Notification.Lark.WebhookURL = ""
		err := validator.Validate(cfg)
		assert.NoError(t, err)

		// 测试启用通知但WebhookURL为空
		cfg = createValidConfig()
		cfg.Notification.Enabled = true
		cfg.Notification.Lark.WebhookURL = ""
		err = validator.Validate(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "飞书Webhook URL不能为空")

		// 测试非HTTPS的WebhookURL
		cfg = createValidConfig()
		cfg.Notification.Lark.WebhookURL = "http://example.com"
		err = validator.Validate(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "飞书Webhook URL必须使用HTTPS")

		// 测试默认超时设置
		cfg = createValidConfig()
		cfg.Notification.Lark.Timeout = 0
		err = validator.Validate(cfg)
		assert.NoError(t, err)
		assert.Equal(t, 10*time.Second, cfg.Notification.Lark.Timeout)
	})

	// 测试应用配置验证
	t.Run("应用配置验证", func(t *testing.T) {
		// 测试默认值设置
		cfg := createValidConfig()
		cfg.App.Name = ""
		cfg.App.Version = ""
		cfg.App.Environment = ""
		cfg.App.WorkDir = ""
		cfg.App.DataDir = ""
		err := validator.Validate(cfg)
		assert.NoError(t, err)
		assert.Equal(t, "daily-commit-summarizer", cfg.App.Name)
		assert.Equal(t, "1.0.0", cfg.App.Version)
		assert.Equal(t, "production", cfg.App.Environment)
		assert.Equal(t, ".", cfg.App.WorkDir)
		assert.Equal(t, "./data", cfg.App.DataDir)

		// 测试无效的环境设置
		cfg = createValidConfig()
		cfg.App.Environment = "invalid"
		err = validator.Validate(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "无效的环境配置")
	})

	// 测试日志配置验证
	t.Run("日志配置验证", func(t *testing.T) {
		// 测试无效的日志级别
		cfg := createValidConfig()
		cfg.Logger.Level = "invalid"
		err := validator.Validate(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "无效的日志级别")

		// 测试无效的日志格式
		cfg = createValidConfig()
		cfg.Logger.Format = "invalid"
		err = validator.Validate(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "无效的日志格式")

		// 测试无效的日志输出
		cfg = createValidConfig()
		cfg.Logger.Output = "invalid"
		err = validator.Validate(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "无效的日志输出")

		// 测试文件输出但路径为空
		cfg = createValidConfig()
		cfg.Logger.Output = "file"
		cfg.Logger.FilePath = ""
		err = validator.Validate(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "使用文件输出时，文件路径不能为空")

		// 测试默认值设置
		cfg = createValidConfig()
		cfg.Logger.MaxSize = 0
		cfg.Logger.MaxBackups = -1
		cfg.Logger.MaxAge = -1
		err = validator.Validate(cfg)
		assert.NoError(t, err)
		assert.Equal(t, 100, cfg.Logger.MaxSize)
		assert.Equal(t, 3, cfg.Logger.MaxBackups)
		assert.Equal(t, 7, cfg.Logger.MaxAge)
	})
}

// 创建一个有效的配置用于测试
func createValidConfig() *Config {
	return &Config{
		Git: GitConfig{
			RepoPath:        ".", // 使用当前目录作为有效路径
			RemoteName:      "origin",
			DefaultBranch:   "main",
			ExcludePatterns: []string{"*.lock", "*.log"},
			MaxCommits:      100,
		},
		LLM: LLMConfig{
			APIKey:      "test-api-key",
			BaseURL:     "https://api.openai.com/v1",
			Model:       "gpt-3.5-turbo",
			MaxTokens:   4000,
			Temperature: 0.7,
			Timeout:     30 * time.Second,
			RetryCount:  3,
		},
		Notification: NotificationConfig{
			Lark: LarkConfig{
				WebhookURL: "https://webhook.test.com",
				Secret:     "test-secret",
				Timeout:    10 * time.Second,
			},
			Enabled: true,
		},
		App: AppConfig{
			Name:        "daily-commit-summarizer",
			Version:     "1.0.0",
			Environment: "production",
			Debug:       false,
			WorkDir:     ".",
			DataDir:     "./data",
		},
		Logger: LoggerConfig{
			Level:      "info",
			Format:     "json",
			Output:     "stdout",
			FilePath:   "./logs/app.log",
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     7,
			Compress:   true,
		},
	}
}