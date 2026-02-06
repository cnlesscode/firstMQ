package server

import (
	"net"
	"sync"

	"github.com/cnlesscode/gotool"
)

type SubscribeClient struct {
	Conn     *net.Conn
	Messages chan []byte
	Topic    string
	IP       string
	Status   bool
}

func NewASubscribeClient(conn net.Conn, topic, ip string) *SubscribeClient {
	subscribeClient := &SubscribeClient{
		Conn:     &conn,
		Messages: make(chan []byte, 10000),
		Topic:    topic,
		IP:       ip,
		Status:   true,
	}
	// 启动一个协程从通道内发送消息
	go func(_subscribeClient *SubscribeClient) {
		for {
			msg := <-_subscribeClient.Messages
			_conn := *_subscribeClient.Conn
			err := gotool.WriteTCPResponse(_conn, msg)
			// 如果发送消息失败，则关闭连接
			if err != nil {
				close(_subscribeClient.Messages)
				_conn.Close()
				_subscribeClient.Status = false
				break
			}
		}
	}(subscribeClient)
	// 返回监听客户端
	return subscribeClient
}

func (sc *SubscribeClient) WriteMessage(message []byte) {
	go func() {
		sc.Messages <- message
	}()
}

/*
订阅客户端连接池
subscribeClients map[话题]map[订阅者IP]chan *SubscribeClient
*/
var subscribeClientsMutex sync.RWMutex = sync.RWMutex{}
var subscribeClients map[string]map[string]chan *SubscribeClient = make(map[string]map[string]chan *SubscribeClient)

// TCPServer TCP服务器结构
type TCPServer struct {
	listener net.Listener
}
