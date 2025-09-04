package notification

// LarkMessage 表示飞书消息结构
type LarkMessage struct {
	MsgType string      `json:"msg_type"`
	Content LarkContent `json:"content"`
}

// LarkContent 表示飞书消息内容
type LarkContent struct {
	Text string `json:"text"`
}

// LarkResponse 表示飞书API响应
type LarkResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}