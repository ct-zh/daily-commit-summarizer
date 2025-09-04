// internal/llm/interfaces.go
package llm

import "context"

// Client 定义LLM客户端接口
type Client interface {
	// Chat 发送提示词到LLM API并获取响应
	// prompt: 提示词内容
	// 返回LLM生成的回复内容和可能的错误
	Chat(prompt string) (string, error)

	// ChatWithContext 带上下文的Chat调用，支持取消和超时
	// ctx: 上下文，用于取消和超时控制
	// prompt: 提示词内容
	// 返回LLM生成的回复内容和可能的错误
	ChatWithContext(ctx context.Context, prompt string) (string, error)
}

// ClientProvider 定义LLM客户端提供者接口
type ClientProvider interface {
	// GetClient 获取LLM客户端
	// 返回LLM客户端实例
	GetClient() Client
}