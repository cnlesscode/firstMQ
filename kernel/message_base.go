package kernel

// 消息结构体
type SendMessageStruct struct {
	Action        int
	Topic         string
	ConsumerGroup string
	Data          any
}

// 响应消息结构体
type ResponseMessage struct {
	ErrCode int    `json:"errcode"`
	Data    string `json:"data"`
}

// 消息存储管道
var MessageChannels = make(map[string]chan []byte)
