package kernel

import (
	"net"
)

// 消息存储管道
var MessageChannels = make(map[string]chan []byte)

// 消费消息缓存管道
var ConsumeMessageChannels = make(map[string]*ConsumeMessagesChannel)

// ServerFinder tcp 通信连接
var serverFinderConn net.Conn
