// internal/llm/client.go
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"daily-commit-summarizer/internal/config"
	"daily-commit-summarizer/internal/utils"
)

// OpenAIClient 处理与OpenAI API的交互
type OpenAIClient struct {
	options    *ClientOptions
	httpClient *http.Client
}

// NewOpenAIClient 创建OpenAI客户端
func NewOpenAIClient(options *ClientOptions) Client {
	// 如果没有提供选项，使用默认选项
	if options == nil {
		options = DefaultClientOptions()
	}

	return &OpenAIClient{
		options: options,
		httpClient: &http.Client{
			Timeout: options.Timeout,
		},
	}
}

// NewClientFromConfig 从配置创建客户端
func NewClientFromConfig(config *config.Config) Client {
	options := &ClientOptions{
		APIKey:      config.LLM.APIKey,
		BaseURL:     config.LLM.BaseURL,
		Model:       config.LLM.Model,
		MaxTokens:   config.LLM.MaxTokens,
		Temperature: config.LLM.Temperature,
		Timeout:     config.LLM.Timeout,
		RetryCount:  config.LLM.RetryCount,
		RetryDelay:  time.Second, // 默认重试延迟1秒
	}

	return NewOpenAIClient(options)
}

// Chat 发送提示词到OpenAI API并获取响应
func (oc *OpenAIClient) Chat(prompt string) (string, error) {
	return oc.ChatWithContext(context.Background(), prompt)
}

// ChatWithContext 带上下文的Chat调用，支持取消和超时控制
func (oc *OpenAIClient) ChatWithContext(ctx context.Context, prompt string) (string, error) {
	// 使用重试机制
	result, err := utils.WithRetry(
		func() (interface{}, error) {
			return oc.doChat(ctx, prompt)
		},
		oc.options.RetryCount,
		oc.options.RetryDelay,
	)

	if err != nil {
		return "", err
	}

	return result.(string), nil
}

// doChat 执行实际的API调用
func (oc *OpenAIClient) doChat(ctx context.Context, prompt string) (string, error) {
	payload := ChatPayload{
		Model: oc.options.Model,
		Messages: []ChatMessage{
			{Role: "user", Content: prompt},
		},
		Temperature: oc.options.Temperature,
		MaxTokens:   oc.options.MaxTokens,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("序列化请求负载失败: %w", err)
	}

	// 解析API URL
	baseURL, err := url.Parse(oc.options.BaseURL)
	if err != nil {
		return "", fmt.Errorf("解析API URL失败: %w", err)
	}

	// 构建请求路径
	var path string
	if strings.Contains(oc.options.BaseURL, "azure") {
		// Azure OpenAI API路径格式
		path = fmt.Sprintf("/openai/deployments/%s/chat/completions?api-version=2024-12-01-preview", oc.options.Model)
	} else {
		// 标准OpenAI API路径格式
		path = "/v1/chat/completions"
	}
	baseURL.Path = path

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", baseURL.String(), bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+oc.options.APIKey)

	// 发送请求
	resp, err := oc.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	// 检查响应状态
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("OpenAI HTTP %d: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var response ChatResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	// 提取内容
	if len(response.Choices) > 0 {
		return strings.TrimSpace(response.Choices[0].Message.Content), nil
	}

	return "", fmt.Errorf("响应中没有内容")
}