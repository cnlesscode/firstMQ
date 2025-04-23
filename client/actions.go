package client

// 创建话题
func (mqPool *MQPool) CreateTopic(topic string) (ResponseMessage, error) {
	return mqPool.Send(Message{Action: 3, Topic: topic})
}

// 生产消息
func (mqPool *MQPool) Product(topic string, data []byte) (ResponseMessage, error) {
	return mqPool.Send(Message{
		Action: 1,
		Topic:  topic,
		Data:   data,
	})
}

// 消费消息
func (mqPool *MQPool) Consume(topic, ConsumerGroup string) (ResponseMessage, error) {
	return mqPool.Send(Message{
		Action:        2,
		Topic:         topic,
		ConsumerGroup: ConsumerGroup,
	})
}

// 创建消费者组
func (mqPool *MQPool) CreateConsumerGroup(topic, ConsumerGroup string) (ResponseMessage, error) {
	return mqPool.Send(Message{
		Action:        7,
		Topic:         topic,
		ConsumerGroup: ConsumerGroup,
	})
}

// 获取话题列表
func (mqPool *MQPool) GetTopicList() (ResponseMessage, error) {
	return mqPool.Send(Message{
		Action: 4,
	})
}
