package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// DefaultConfigValidator 默认配置验证器
type DefaultConfigValidator struct{}

// Validate 验证配置
func (v *DefaultConfigValidator) Validate(config *Config) error {
	if config == nil {
		return fmt.Errorf("配置不能为空")
	}

	// 验证Git配置
	if err := v.validateGitConfig(&config.Git); err != nil {
		return fmt.Errorf("Git配置验证失败: %w", err)
	}

	// 验证LLM配置
	if err := v.validateLLMConfig(&config.LLM); err != nil {
		return fmt.Errorf("LLM配置验证失败: %w", err)
	}

	// 验证通知配置
	if err := v.validateNotificationConfig(&config.Notification); err != nil {
		return fmt.Errorf("通知配置验证失败: %w", err)
	}

	// 验证应用配置
	if err := v.validateAppConfig(&config.App); err != nil {
		return fmt.Errorf("应用配置验证失败: %w", err)
	}

	// 验证日志配置
	if err := v.validateLoggerConfig(&config.Logger); err != nil {
		return fmt.Errorf("日志配置验证失败: %w", err)
	}

	return nil
}

// validateGitConfig 验证Git配置
func (v *DefaultConfigValidator) validateGitConfig(config *GitConfig) error {
	if config.RepoPath == "" {
		return fmt.Errorf("仓库路径不能为空")
	}

	// 检查仓库路径是否存在
	// 在测试环境中跳过路径存在性检查
	if os.Getenv("GO_ENV") != "test" {
		if _, err := os.Stat(config.RepoPath); os.IsNotExist(err) {
			return fmt.Errorf("仓库路径不存在: %s", config.RepoPath)
		}
	}

	if config.RemoteName == "" {
		config.RemoteName = "origin"
	}

	if config.DefaultBranch == "" {
		config.DefaultBranch = "main"
	}

	if config.MaxCommits <= 0 {
		config.MaxCommits = 100
	}

	return nil
}

// validateLLMConfig 验证LLM配置
func (v *DefaultConfigValidator) validateLLMConfig(config *LLMConfig) error {
	if config.APIKey == "" {
		return fmt.Errorf("LLM API密钥不能为空")
	}

	if config.BaseURL == "" {
		config.BaseURL = "https://api.openai.com/v1"
	}

	if config.Model == "" {
		config.Model = "gpt-3.5-turbo"
	}

	if config.MaxTokens <= 0 {
		config.MaxTokens = 4000
	}

	if config.Temperature < 0 || config.Temperature > 2 {
		return fmt.Errorf("温度参数必须在0-2之间")
	}

	if config.Timeout <= 0 {
		config.Timeout = 30 * time.Second
	}

	if config.RetryCount < 0 {
		config.RetryCount = 3
	}

	return nil
}

// validateNotificationConfig 验证通知配置
func (v *DefaultConfigValidator) validateNotificationConfig(config *NotificationConfig) error {
	if !config.Enabled {
		return nil
	}

	// 验证飞书配置
	if err := v.validateLarkConfig(&config.Lark); err != nil {
		return fmt.Errorf("飞书配置验证失败: %w", err)
	}

	return nil
}

// validateLarkConfig 验证飞书配置
func (v *DefaultConfigValidator) validateLarkConfig(config *LarkConfig) error {
	if config.WebhookURL == "" {
		return fmt.Errorf("飞书Webhook URL不能为空")
	}

	if !strings.HasPrefix(config.WebhookURL, "https://") {
		return fmt.Errorf("飞书Webhook URL必须使用HTTPS")
	}

	if config.Timeout <= 0 {
		config.Timeout = 10 * time.Second
	}

	return nil
}

// validateAppConfig 验证应用配置
func (v *DefaultConfigValidator) validateAppConfig(config *AppConfig) error {
	if config.Name == "" {
		config.Name = "daily-commit-summarizer"
	}

	if config.Version == "" {
		config.Version = "1.0.0"
	}

	if config.Environment == "" {
		config.Environment = "production"
	}

	validEnvs := []string{"development", "testing", "staging", "production"}
	validEnv := false
	for _, env := range validEnvs {
		if config.Environment == env {
			validEnv = true
			break
		}
	}
	if !validEnv {
		return fmt.Errorf("无效的环境配置: %s，支持的环境: %v", config.Environment, validEnvs)
	}

	if config.WorkDir == "" {
		config.WorkDir = "."
	}

	if config.DataDir == "" {
		config.DataDir = "./data"
	}

	return nil
}

// validateLoggerConfig 验证日志配置
func (v *DefaultConfigValidator) validateLoggerConfig(config *LoggerConfig) error {
	validLevels := []string{"debug", "info", "warn", "error", "fatal"}
	validLevel := false
	for _, level := range validLevels {
		if strings.ToLower(config.Level) == level {
			config.Level = level
			validLevel = true
			break
		}
	}
	if !validLevel {
		return fmt.Errorf("无效的日志级别: %s，支持的级别: %v", config.Level, validLevels)
	}

	validFormats := []string{"json", "text"}
	validFormat := false
	for _, format := range validFormats {
		if strings.ToLower(config.Format) == format {
			config.Format = format
			validFormat = true
			break
		}
	}
	if !validFormat {
		return fmt.Errorf("无效的日志格式: %s，支持的格式: %v", config.Format, validFormats)
	}

	validOutputs := []string{"stdout", "file"}
	validOutput := false
	for _, output := range validOutputs {
		if strings.ToLower(config.Output) == output {
			config.Output = output
			validOutput = true
			break
		}
	}
	if !validOutput {
		return fmt.Errorf("无效的日志输出: %s，支持的输出: %v", config.Output, validOutputs)
	}

	if config.Output == "file" && config.FilePath == "" {
		return fmt.Errorf("使用文件输出时，文件路径不能为空")
	}

	if config.MaxSize <= 0 {
		config.MaxSize = 100
	}

	if config.MaxBackups < 0 {
		config.MaxBackups = 3
	}

	if config.MaxAge < 0 {
		config.MaxAge = 7
	}

	return nil
}