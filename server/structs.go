package server

import (
	"net"
	"sync"

	"github.com/cnlesscode/gotool"
)

type SubscribeClient struct {
	Id       string
	Conn     *net.Conn
	Messages chan []byte
	Topic    string
	IP       string
}

func NewASubscribeClient(conn net.Conn, topic, ip, id string) *SubscribeClient {
	subscribeClient := &SubscribeClient{
		Id:       id,
		Conn:     &conn,
		Messages: make(chan []byte, 100000),
		Topic:    topic,
		IP:       ip,
	}
	// 启动一个协程从通道内发送消息
	go func(_subscribeClient *SubscribeClient) {
		for {
			if _subscribeClient == nil {
				break
			}
			msg := <-_subscribeClient.Messages
			err := gotool.WriteTCPResponse(*_subscribeClient.Conn, msg)
			// 如果发送消息失败，则关闭连接
			if err != nil {
				close(_subscribeClient.Messages)
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
subscribeClients map[话题]map[订阅者ID]*SubscribeClient
*/
var subscribeClientsMutex sync.RWMutex = sync.RWMutex{}
var subscribeClients map[string]map[string]*SubscribeClient = make(map[string]map[string]*SubscribeClient)
