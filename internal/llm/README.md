# LLM模块 (Large Language Model)

本模块提供了与大型语言模型（如OpenAI GPT系列）交互的功能，用于生成提交摘要和日报内容。

## 模块结构

```
internal/llm/
├── interfaces.go    # 接口定义
├── models.go        # 数据模型
├── client.go        # 客户端实现
├── client_test.go   # 测试文件
└── README.md        # 本文档
```

## 核心接口

### Client

`Client` 接口定义了与LLM交互的核心方法：

```go
type Client interface {
	// Chat 发送提示词到LLM API并获取响应
	Chat(prompt string) (string, error)

	// ChatWithContext 带上下文的Chat调用，支持取消和超时
	ChatWithContext(ctx context.Context, prompt string) (string, error)
}
```

### ClientProvider

`ClientProvider` 接口用于获取LLM客户端实例：

```go
type ClientProvider interface {
	// GetClient 获取LLM客户端
	GetClient() Client
}
```

## 数据模型

### ChatMessage

表示聊天消息：

```go
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
```

### ChatPayload

表示LLM API请求负载：

```go
type ChatPayload struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
}
```

### ClientOptions

表示LLM客户端选项：

```go
type ClientOptions struct {
	APIKey      string
	BaseURL     string
	Model       string
	MaxTokens   int
	Temperature float64
	Timeout     time.Duration
	RetryCount  int
	RetryDelay  time.Duration
}
```

## 使用示例

### 基本使用

```go
package main

import (
	"fmt"
	"log"

	"daily-commit-summarizer/internal/llm"
)

func main() {
	// 创建客户端选项
	options := &llm.ClientOptions{
		APIKey:      "your-api-key",
		BaseURL:     "https://api.openai.com/v1",
		Model:       "gpt-3.5-turbo",
		MaxTokens:   4000,
		Temperature: 0.7,
		Timeout:     30 * time.Second,
		RetryCount:  3,
		RetryDelay:  1 * time.Second,
	}

	// 创建客户端
	client := llm.NewOpenAIClient(options)

	// 发送提示词
	response, err := client.Chat("请总结以下代码变更的主要内容：...")
	if err != nil {
		log.Fatalf("调用LLM API失败: %v", err)
	}

	fmt.Println("LLM回复:", response)
}
```

### 使用配置创建客户端

```go
package main

import (
	"fmt"
	"log"

	"daily-commit-summarizer/internal/config"
	"daily-commit-summarizer/internal/llm"
)

func main() {
	// 加载配置
	configLoader := config.NewDefaultConfigLoader("./configs/config.yaml")
	cfg, err := configLoader.Load(context.Background())
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 从配置创建客户端
	client := llm.NewClientFromConfig(cfg)

	// 发送提示词
	response, err := client.Chat("请总结以下代码变更的主要内容：...")
	if err != nil {
		log.Fatalf("调用LLM API失败: %v", err)
	}

	fmt.Println("LLM回复:", response)
}
```

### 使用上下文控制

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"daily-commit-summarizer/internal/llm"
)

func main() {
	// 创建客户端
	client := llm.NewOpenAIClient(llm.DefaultClientOptions())

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 发送提示词
	response, err := client.ChatWithContext(ctx, "请总结以下代码变更的主要内容：...")
	if err != nil {
		log.Fatalf("调用LLM API失败: %v", err)
	}

	fmt.Println("LLM回复:", response)
}
```

## 错误处理

LLM模块提供了详细的错误信息，包括：

- API请求错误（网络问题、超时等）
- API响应错误（状态码非2xx）
- 响应解析错误
- 上下文取消错误

示例：

```go
response, err := client.Chat(prompt)
if err != nil {
	switch {
	case strings.Contains(err.Error(), "timeout"):
		// 处理超时错误
		fmt.Println("请求超时，稍后重试")
	case strings.Contains(err.Error(), "OpenAI HTTP 429"):
		// 处理速率限制错误
		fmt.Println("请求过于频繁，请稍后重试")
	case strings.Contains(err.Error(), "OpenAI HTTP 401"):
		// 处理认证错误
		fmt.Println("API密钥无效或已过期")
	default:
		// 处理其他错误
		fmt.Printf("发生错误: %v\n", err)
	}
	return
}
```

## 重试机制

LLM模块内置了重试机制，可以通过`ClientOptions`中的`RetryCount`和`RetryDelay`参数进行配置：

- `RetryCount`: 重试次数（默认3次）
- `RetryDelay`: 重试间隔时间（默认1秒）

重试机制会在以下情况下触发：

- 网络错误
- 服务器错误（5xx状态码）
- 速率限制错误（429状态码）

## 测试

运行测试：

```bash
GO111MODULE=on go test ./internal/llm/... -v
```

生成测试覆盖率报告：

```bash
GO111MODULE=on go test ./internal/llm/... -v -cover
```

## 注意事项

1. **API密钥安全**: 不要在代码中硬编码API密钥，应该通过环境变量或配置文件安全地传递
2. **错误处理**: 始终检查并处理返回的错误
3. **超时控制**: 对于重要的请求，使用`ChatWithContext`方法设置超时
4. **并发限制**: 注意OpenAI API的速率限制，避免过于频繁的请求
5. **内容安全**: 确保发送给API的内容符合使用条款和内容政策

## 依赖

- 标准库: `context`, `encoding/json`, `net/http`, `time`等
- 内部模块: `config`, `utils`

## 版本历史

- v1.0.0: 初始版本，支持OpenAI API调用和重试机制