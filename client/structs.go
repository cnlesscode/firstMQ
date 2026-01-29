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
	Conn   net.Conn // TCP 连接指针
	Addr   string   // TCP 服务 Address
	Status bool     // 状态
}

// 接池结构体
type MQPool struct {
	ServerFindAddr  string // ServerFind 地址
	Addresses       map[string]int
	Connections     chan *MQConnection // 总连接池，存放所有连接
	BadConnections  chan *MQConnection // 坏的的连接池
	CapacityForNode int                // 每个节点连接池总容量
	ErrorMessage    chan []byte        // 错误消息临时记录通道
	ServerStatus    map[string]bool    // 服务器状态
}
