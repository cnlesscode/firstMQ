package server

// 将消息转发给订阅者
func SendMessageToSubscribers(message ReceiveMessage) {
	// 获取订阅者列表
	clients := subscribeClients[message.Topic]
	if len(clients) < 1 {
		return
	}
	// 遍历订阅者列表
	for _, client := range clients {
		client.WriteMessage(message.Data)
	}
}
