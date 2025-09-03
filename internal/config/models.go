package config

import "time"

// Config 应用配置结构
type Config struct {
	// Git 配置
	Git GitConfig `json:"git" yaml:"git"`
	// LLM 配置
	LLM LLMConfig `json:"llm" yaml:"llm"`
	// 通知配置
	Notification NotificationConfig `json:"notification" yaml:"notification"`
	// 应用配置
	App AppConfig `json:"app" yaml:"app"`
	// 日志配置
	Logger LoggerConfig `json:"logger" yaml:"logger"`
}

// GitConfig Git相关配置
type GitConfig struct {
	// 仓库路径
	RepoPath string `json:"repo_path" yaml:"repo_path"`
	// 远程仓库名称
	RemoteName string `json:"remote_name" yaml:"remote_name"`
	// 默认分支
	DefaultBranch string `json:"default_branch" yaml:"default_branch"`
	// 排除的文件模式
	ExcludePatterns []string `json:"exclude_patterns" yaml:"exclude_patterns"`
	// 最大提交数量
	MaxCommits int `json:"max_commits" yaml:"max_commits"`
}

// LLMConfig LLM相关配置
type LLMConfig struct {
	// API密钥
	APIKey string `json:"api_key" yaml:"api_key"`
	// API基础URL
	BaseURL string `json:"base_url" yaml:"base_url"`
	// 模型名称
	Model string `json:"model" yaml:"model"`
	// 最大令牌数
	MaxTokens int `json:"max_tokens" yaml:"max_tokens"`
	// 温度参数
	Temperature float64 `json:"temperature" yaml:"temperature"`
	// 请求超时时间
	Timeout time.Duration `json:"timeout" yaml:"timeout"`
	// 重试次数
	RetryCount int `json:"retry_count" yaml:"retry_count"`
}

// NotificationConfig 通知相关配置
type NotificationConfig struct {
	// 飞书配置
	Lark LarkConfig `json:"lark" yaml:"lark"`
	// 是否启用通知
	Enabled bool `json:"enabled" yaml:"enabled"`
}

// LarkConfig 飞书通知配置
type LarkConfig struct {
	// Webhook URL
	WebhookURL string `json:"webhook_url" yaml:"webhook_url"`
	// 签名密钥
	Secret string `json:"secret" yaml:"secret"`
	// 超时时间
	Timeout time.Duration `json:"timeout" yaml:"timeout"`
}

// AppConfig 应用相关配置
type AppConfig struct {
	// 应用名称
	Name string `json:"name" yaml:"name"`
	// 应用版本
	Version string `json:"version" yaml:"version"`
	// 环境
	Environment string `json:"environment" yaml:"environment"`
	// 调试模式
	Debug bool `json:"debug" yaml:"debug"`
	// 工作目录
	WorkDir string `json:"work_dir" yaml:"work_dir"`
	// 数据目录
	DataDir string `json:"data_dir" yaml:"data_dir"`
}

// LoggerConfig 日志相关配置
type LoggerConfig struct {
	// 日志级别
	Level string `json:"level" yaml:"level"`
	// 日志格式 (json/text)
	Format string `json:"format" yaml:"format"`
	// 输出目标 (stdout/file)
	Output string `json:"output" yaml:"output"`
	// 日志文件路径
	FilePath string `json:"file_path" yaml:"file_path"`
	// 最大文件大小(MB)
	MaxSize int `json:"max_size" yaml:"max_size"`
	// 最大备份数量
	MaxBackups int `json:"max_backups" yaml:"max_backups"`
	// 最大保留天数
	MaxAge int `json:"max_age" yaml:"max_age"`
	// 是否压缩
	Compress bool `json:"compress" yaml:"compress"`
}