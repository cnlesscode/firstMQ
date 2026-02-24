package client

import "github.com/cnlesscode/firstMQ/server"

// 创建话题
func (mqPool *MQPool) CreateTopic(topic string) (ResponseMessage, error) {
	return mqPool.Send(
		Message{
			Action: server.CreateTopic,
			Topic:  topic,
		})
}

// 生产消息
func (mqPool *MQPool) Product(topic string, data []byte, Broadcast bool) (ResponseMessage, error) {
	return mqPool.Send(
		Message{
			Action:    server.Product,
			Topic:     topic,
			Data:      data,
			Broadcast: Broadcast,
		})
}

// 消费消息
func (mqPool *MQPool) Consume(topic, consumerGroup string) (ResponseMessage, error) {
	return mqPool.Send(
		Message{
			Action:        server.Consume,
			Topic:         topic,
			ConsumerGroup: consumerGroup,
		})
}

// 创建消费者组
func (mqPool *MQPool) CreateConsumerGroup(topic, consumerGroup string) (ResponseMessage, error) {
	return mqPool.Send(
		Message{
			Action:        server.CreateConsumeGroup,
			Topic:         topic,
			ConsumerGroup: consumerGroup,
		})
}

// 获取话题列表
func (mqPool *MQPool) GetTopicList() (ResponseMessage, error) {
	return mqPool.Send(
		Message{
			Action: server.TopicList,
		})
}

// Ping
func (mqPool *MQPool) Ping() (ResponseMessage, error) {
	return mqPool.Send(
		Message{
			Action: server.Ping,
		},
	)
}

// Broadcast
func (mqPool *MQPool) Broadcast(topic string, msg []byte) (ResponseMessage, error) {
	return mqPool.Send(
		Message{
			Action: server.Broadcast,
			Topic:  topic,
			Data:   msg,
		},
	)
}
