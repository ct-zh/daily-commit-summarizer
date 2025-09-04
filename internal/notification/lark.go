package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"daily-commit-summarizer/internal/config"
	"daily-commit-summarizer/internal/utils"
)

// LarkNotifier 处理飞书通知
type LarkNotifier struct {
	config     *config.Config
	httpClient *http.Client
}

// NewLarkNotifier 创建飞书通知器
func NewLarkNotifier(config *config.Config) Notifier {
	return &LarkNotifier{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Notification.Lark.Timeout,
		},
	}
}

// Send 发送消息到飞书
func (ln *LarkNotifier) Send(text string) error {
	if ln.config.Notification.Lark.WebhookURL == "" {
		fmt.Printf("LARK_WEBHOOK_URL 未配置，以下为最终日报文本：\n\n%s\n", text)
		return nil
	}

	// 使用重试机制
	_, err := utils.WithRetry(
		func() (interface{}, error) {
			return nil, ln.doSend(text)
		},
		3, // 重试3次
		1*time.Second, // 重试间隔1秒
	)

	return err
}

// doSend 执行实际的发送操作
func (ln *LarkNotifier) doSend(text string) error {
	// 构建请求负载
	payload := map[string]interface{}{
		"msg_type": "text",
		"content": map[string]string{
			"text": text,
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("序列化请求负载失败: %w", err)
	}

	// 解析Webhook URL
	webhookURL, err := url.Parse(ln.config.Notification.Lark.WebhookURL)
	if err != nil {
		return fmt.Errorf("解析Webhook URL失败: %w", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequest("POST", webhookURL.String(), bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := ln.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("发送HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("飞书Webhook HTTP %d", resp.StatusCode)
	}

	return nil
}