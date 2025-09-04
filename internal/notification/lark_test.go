package notification

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"daily-commit-summarizer/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLarkNotifier_Send(t *testing.T) {
	tests := []struct {
		name           string
		webhookURL     string
		text           string
		serverResponse int
		expectError    bool
	}{
		{
			name:           "成功发送消息",
			webhookURL:     "http://test.webhook.url",
			text:           "测试消息",
			serverResponse: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "服务器返回错误",
			webhookURL:     "http://test.webhook.url",
			text:           "测试消息",
			serverResponse: http.StatusInternalServerError,
			expectError:    true,
		},
		{
			name:           "空的Webhook URL",
			webhookURL:     "",
			text:           "测试消息",
			serverResponse: http.StatusOK,
			expectError:    false, // 应该不报错，只是打印消息
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试服务器
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.serverResponse)
			}))
			defer server.Close()

			// 如果有webhook URL，使用测试服务器的URL
			webhookURL := tt.webhookURL
			if webhookURL != "" && webhookURL != "http://test.webhook.url" {
				webhookURL = server.URL
			} else if webhookURL == "http://test.webhook.url" {
				webhookURL = server.URL
			}

			// 创建配置
			cfg := &config.Config{
				Notification: config.NotificationConfig{
					Lark: config.LarkConfig{
						WebhookURL: webhookURL,
						Timeout:    5 * time.Second,
					},
				},
			}

			// 创建通知器
			notifier := NewLarkNotifier(cfg)

			// 发送消息
			err := notifier.Send(tt.text)

			// 验证结果
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLarkNotifier_doSend(t *testing.T) {
	// 创建测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求方法
		assert.Equal(t, "POST", r.Method)
		// 验证Content-Type
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// 创建配置
	cfg := &config.Config{
		Notification: config.NotificationConfig{
			Lark: config.LarkConfig{
				WebhookURL: server.URL,
				Timeout:    5 * time.Second,
			},
		},
	}

	// 创建通知器
	larkNotifier := NewLarkNotifier(cfg).(*LarkNotifier)

	// 测试doSend方法
	err := larkNotifier.doSend("测试消息")
	require.NoError(t, err)
}

func TestLarkNotifier_doSend_InvalidURL(t *testing.T) {
	// 创建配置，使用无效的URL
	cfg := &config.Config{
		Notification: config.NotificationConfig{
			Lark: config.LarkConfig{
				WebhookURL: "invalid-url",
				Timeout:    5 * time.Second,
			},
		},
	}

	// 创建通知器
	larkNotifier := NewLarkNotifier(cfg).(*LarkNotifier)

	// 测试doSend方法
	err := larkNotifier.doSend("测试消息")
	assert.Error(t, err)
	// 无效URL会导致HTTP请求失败
	assert.Contains(t, err.Error(), "发送HTTP请求失败")
}