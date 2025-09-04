package notification

// Notifier 定义通知器接口
type Notifier interface {
	// Send 发送消息
	Send(text string) error
}