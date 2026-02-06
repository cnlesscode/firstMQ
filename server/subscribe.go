package server

import "time"

// 将消息转发给订阅者
func SendMessageToSubscribers(message ReceiveMessage) {
	// 获取订阅者列表
	clients := subscribeClients[message.Topic]
	// 遍历订阅者列表
	for ip, clientsChannel := range clients {
		go func(_ip string, _message ReceiveMessage, _chan chan *SubscribeClient) {
			var sc *SubscribeClient
			tryTimer := 0
			for {
				select {
				case sc = <-_chan:
				case <-time.After(time.Millisecond * 200):
				}
				if sc != nil {
					if sc.Status {
						break
					} else {
						tryTimer++
						continue
					}
				}
				tryTimer++
				if tryTimer > 5 {
					return
				}
			}
			sc.WriteMessage(_message.Data)
			// 将订阅连接放回连接池
			subscribeClients[_message.Topic][_ip] <- sc
		}(ip, message, clientsChannel)
	}
}
