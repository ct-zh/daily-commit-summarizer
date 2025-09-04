# 通知模块 (Notification)

通知模块负责将生成的日报发送到各种通知平台，目前支持飞书(Lark)通知。

## 功能特性

- **统一接口设计**: 通过 `Notifier` 接口定义统一的通知发送方法
- **飞书集成**: 支持通过 Webhook 发送消息到飞书群聊
- **错误处理**: 完善的错误处理和重试机制
- **配置灵活**: 支持通过配置文件或环境变量配置通知参数
- **测试完备**: 提供完整的单元测试覆盖

## 接口设计

### Notifier 接口

```go
type Notifier interface {
    // Send 发送消息
    Send(text string) error
}
```

### 数据模型

- `LarkMessage`: 飞书消息结构
- `LarkContent`: 飞书消息内容
- `LarkResponse`: 飞书API响应

## 使用示例

### 创建飞书通知器

```go
package main

import (
    "daily-commit-summarizer/internal/config"
    "daily-commit-summarizer/internal/notification"
)

func main() {
    // 加载配置
    cfg := &config.Config{
        Notification: config.NotificationConfig{
            Lark: config.LarkConfig{
                WebhookURL: "https://open.feishu.cn/open-apis/bot/v2/hook/your-webhook-url",
                Timeout:    10 * time.Second,
            },
        },
    }
    
    // 创建通知器
    notifier := notification.NewLarkNotifier(cfg)
    
    // 发送消息
    err := notifier.Send("今日代码提交摘要：\n\n...")
    if err != nil {
        log.Printf("发送通知失败: %v", err)
    }
}
```

## 飞书通知器 (LarkNotifier)

### 特性

1. **Webhook 集成**: 通过飞书机器人 Webhook 发送消息
2. **自动重试**: 内置重试机制，提高发送成功率
3. **超时控制**: 可配置的HTTP请求超时时间
4. **错误处理**: 详细的错误信息和状态码检查
5. **调试模式**: 当未配置 Webhook URL 时，将消息打印到控制台

### 配置要求

在配置文件中需要设置以下参数：

```yaml
notification:
  lark:
    webhook_url: "https://open.feishu.cn/open-apis/bot/v2/hook/your-webhook-url"
    timeout: "10s"
```

或通过环境变量：

```bash
export LARK_WEBHOOK_URL="https://open.feishu.cn/open-apis/bot/v2/hook/your-webhook-url"
```

### 消息格式

飞书通知器发送的消息格式为纯文本，支持换行和基本格式化。消息结构：

```json
{
  "msg_type": "text",
  "content": {
    "text": "消息内容"
  }
}
```

## 扩展性

### 添加新的通知平台

要添加新的通知平台支持，只需：

1. 实现 `Notifier` 接口
2. 在配置中添加相应的配置项
3. 编写相应的单元测试

示例：

```go
type SlackNotifier struct {
    config *config.Config
    // ...
}

func (s *SlackNotifier) Send(text string) error {
    // 实现 Slack 消息发送逻辑
    return nil
}
```

## 测试

运行单元测试：

```bash
go test ./internal/notification/...
```

测试覆盖的场景：
- 成功发送消息
- 服务器返回错误状态码
- 网络连接失败
- 无效的 Webhook URL
- 空的 Webhook URL（调试模式）

## 设计特点

1. **接口驱动**: 通过接口定义统一的通知行为，便于扩展和测试
2. **错误恢复**: 内置重试机制和降级策略
3. **配置外部化**: 支持多种配置方式，便于部署和管理
4. **类型安全**: 使用强类型定义消息结构，避免运行时错误
5. **测试友好**: 提供完整的单元测试和模拟对象

## 依赖关系

- `internal/config`: 配置管理
- `internal/utils`: 工具函数（重试机制）
- `net/http`: HTTP客户端
- `encoding/json`: JSON序列化