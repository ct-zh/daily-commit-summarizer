// internal/llm/models.go
package llm

import "time"

// ChatMessage 表示聊天消息
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatPayload 表示LLM API请求负载
type ChatPayload struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
}

// ChatResponse 表示LLM API响应
type ChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// ClientOptions 表示LLM客户端选项
type ClientOptions struct {
	// API密钥
	APIKey string
	// API基础URL
	BaseURL string
	// 模型名称
	Model string
	// 最大令牌数
	MaxTokens int
	// 温度参数
	Temperature float64
	// 请求超时时间
	Timeout time.Duration
	// 重试次数
	RetryCount int
	// 重试延迟
	RetryDelay time.Duration
}

// DefaultClientOptions 返回默认的客户端选项
func DefaultClientOptions() *ClientOptions {
	return &ClientOptions{
		BaseURL:     "https://api.openai.com/v1",
		Model:       "gpt-3.5-turbo",
		MaxTokens:   4000,
		Temperature: 0.7,
		Timeout:     30 * time.Second,
		RetryCount:  3,
		RetryDelay:  1 * time.Second,
	}
}