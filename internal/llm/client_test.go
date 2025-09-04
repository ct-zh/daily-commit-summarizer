// internal/llm/client_test.go
package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"daily-commit-summarizer/internal/config"
)

// TestNewOpenAIClient 测试创建OpenAI客户端
func TestNewOpenAIClient(t *testing.T) {
	// 测试使用默认选项
	client := NewOpenAIClient(nil)
	assert.NotNil(t, client, "客户端不应为nil")

	// 测试使用自定义选项
	options := &ClientOptions{
		APIKey:      "test-api-key",
		BaseURL:     "https://test-api.com",
		Model:       "test-model",
		MaxTokens:   1000,
		Temperature: 0.5,
		Timeout:     10 * time.Second,
		RetryCount:  2,
		RetryDelay:  500 * time.Millisecond,
	}
	client = NewOpenAIClient(options)
	assert.NotNil(t, client, "客户端不应为nil")

	// 类型断言检查
	_, ok := client.(*OpenAIClient)
	assert.True(t, ok, "客户端应该是OpenAIClient类型")
}

// TestNewClientFromConfig 测试从配置创建客户端
func TestNewClientFromConfig(t *testing.T) {
	// 创建测试配置
	cfg := &config.Config{
		LLM: config.LLMConfig{
			APIKey:      "test-api-key",
			BaseURL:     "https://test-api.com",
			Model:       "test-model",
			MaxTokens:   1000,
			Temperature: 0.5,
			Timeout:     10 * time.Second,
			RetryCount:  2,
		},
	}

	client := NewClientFromConfig(cfg)
	assert.NotNil(t, client, "客户端不应为nil")

	// 类型断言检查
	_, ok := client.(*OpenAIClient)
	assert.True(t, ok, "客户端应该是OpenAIClient类型")
}

// TestOpenAIClient_Chat 测试Chat方法
func TestOpenAIClient_Chat(t *testing.T) {
	// 创建模拟服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求方法
		assert.Equal(t, "POST", r.Method)

		// 验证请求头
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))

		// 解析请求体
		var payload ChatPayload
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&payload)
		assert.NoError(t, err)

		// 验证请求负载
		assert.Equal(t, "test-model", payload.Model)
		assert.Len(t, payload.Messages, 1)
		assert.Equal(t, "user", payload.Messages[0].Role)
		assert.Equal(t, "测试提示词", payload.Messages[0].Content)

		// 返回模拟响应
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := `{
			"id": "test-id",
			"object": "chat.completion",
			"created": 1677858242,
			"model": "test-model",
			"choices": [
				{
					"index": 0,
					"message": {
						"role": "assistant",
						"content": "测试回复内容"
					},
					"finish_reason": "stop"
				}
			],
			"usage": {
				"prompt_tokens": 10,
				"completion_tokens": 20,
				"total_tokens": 30
			}
		}`
		w.Write([]byte(response))
	}))
	defer server.Close()

	// 创建客户端
	options := &ClientOptions{
		APIKey:      "test-api-key",
		BaseURL:     server.URL,
		Model:       "test-model",
		MaxTokens:   1000,
		Temperature: 0.5,
		Timeout:     10 * time.Second,
		RetryCount:  2,
		RetryDelay:  500 * time.Millisecond,
	}
	client := NewOpenAIClient(options)

	// 测试Chat方法
	response, err := client.Chat("测试提示词")
	assert.NoError(t, err)
	assert.Equal(t, "测试回复内容", response)
}

// TestOpenAIClient_ChatWithContext 测试ChatWithContext方法
func TestOpenAIClient_ChatWithContext(t *testing.T) {
	// 创建模拟服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 返回模拟响应
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := `{
			"id": "test-id",
			"object": "chat.completion",
			"created": 1677858242,
			"model": "test-model",
			"choices": [
				{
					"index": 0,
					"message": {
						"role": "assistant",
						"content": "测试回复内容"
					},
					"finish_reason": "stop"
				}
			],
			"usage": {
				"prompt_tokens": 10,
				"completion_tokens": 20,
				"total_tokens": 30
			}
		}`
		w.Write([]byte(response))
	}))
	defer server.Close()

	// 创建客户端
	options := &ClientOptions{
		APIKey:      "test-api-key",
		BaseURL:     server.URL,
		Model:       "test-model",
		MaxTokens:   1000,
		Temperature: 0.5,
		Timeout:     10 * time.Second,
		RetryCount:  2,
		RetryDelay:  500 * time.Millisecond,
	}
	client := NewOpenAIClient(options)

	// 测试带上下文的Chat方法
	ctx := context.Background()
	response, err := client.ChatWithContext(ctx, "测试提示词")
	assert.NoError(t, err)
	assert.Equal(t, "测试回复内容", response)

	// 测试取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 立即取消
	_, err = client.ChatWithContext(ctx, "测试提示词")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
}

// TestOpenAIClient_ChatError 测试Chat错误处理
func TestOpenAIClient_ChatError(t *testing.T) {
	// 创建返回错误的模拟服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":{"message":"测试错误","type":"invalid_request_error"}}`))  
	}))
	defer server.Close()

	// 创建客户端
	options := &ClientOptions{
		APIKey:      "test-api-key",
		BaseURL:     server.URL,
		Model:       "test-model",
		MaxTokens:   1000,
		Temperature: 0.5,
		Timeout:     10 * time.Second,
		RetryCount:  1, // 只重试一次以加快测试
		RetryDelay:  100 * time.Millisecond,
	}
	client := NewOpenAIClient(options)

	// 测试错误处理
	_, err := client.Chat("测试提示词")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "OpenAI HTTP 400")
}

// TestOpenAIClient_EmptyResponse 测试空响应处理
func TestOpenAIClient_EmptyResponse(t *testing.T) {
	// 创建返回空响应的模拟服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := `{"id":"test-id","object":"chat.completion","created":1677858242,"model":"test-model","choices":[],"usage":{"prompt_tokens":10,"completion_tokens":0,"total_tokens":10}}`
		w.Write([]byte(response))
	}))
	defer server.Close()

	// 创建客户端
	options := &ClientOptions{
		APIKey:      "test-api-key",
		BaseURL:     server.URL,
		Model:       "test-model",
		MaxTokens:   1000,
		Temperature: 0.5,
		Timeout:     10 * time.Second,
		RetryCount:  1,
		RetryDelay:  100 * time.Millisecond,
	}
	client := NewOpenAIClient(options)

	// 测试空响应处理
	_, err := client.Chat("测试提示词")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "响应中没有内容")
}