package client

import (
	"net"
)

// MQ 消息结构体
type Message struct {
	Action        int
	Topic         string
	ConsumerGroup string
	Data          []byte
}

// 响应消息结构体
type ResponseMessage struct {
	ErrCode int
	Data    string
}

// MQ TCP 连接对象结构体
type MQConnection struct {
	MapKey string
	Conn   net.Conn // TCP 连接指针
	Addr   string   // TCP 服务 Address
	Status bool     // 状态
}

// 总连接池结构体
type MQConnectionPool struct {
	Key                  string
	ServerFindAddr       string // ServerFind 地址
	Addresses            map[string]any
	AddressesLen         int                           // TCP 服务地址数量
	MQConnections        map[string]chan *MQConnection // 对应各服务器的连接池
	AllConnections       chan *MQConnection            // 总连接池，存放所有连接
	Capacity             int                           // 连接池总容量
	ConnNumber           map[string]int                // 对应某个服务器应建立的连接数量
	ConnDifferenceNumber map[string]int                // 连接池数量与实际数量差值
	ErrorMessage         chan []byte                   // 错误消息临时记录通道
	ServerStatus         map[string]bool               // 各服务器节点状态
}
