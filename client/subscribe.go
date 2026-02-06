package client

import (
	"encoding/json"
	"net"
	"time"

	"github.com/cnlesscode/firstMQ/configs"
	"github.com/cnlesscode/firstMQ/server"
	"github.com/cnlesscode/gotool"
	serverFinderClient "github.com/cnlesscode/serverFinder/client"
)

var subscribeServerConnections map[string]*MQConnection = make(map[string]*MQConnection)

// 客户端订阅话题
// 订阅端在服务端以连接池形式存储
func Subscribe(ServerFinderAddr, topicName string, poolSize int, onMessage func(msg []byte)) {
	serverFinderClient.Listen(
		ServerFinderAddr,
		configs.ServerFinderVarKey,
		func(message map[string]int) {
			if len(message) < 1 {
				return
			}
			for k := range message {
				// 创建订阅任务
				for i := 0; i < poolSize; i++ {
					// 创建订阅任务
					go subscribeBase(k, topicName, onMessage)
				}
			}
		},
	)
}

func subscribeBase(mqServerAddr, topicName string, onMessage func(msg []byte)) {
	var keyName = mqServerAddr + "_" + topicName
	if _, ok := subscribeServerConnections[keyName]; ok {
		return
	}
SubscribeLoop:
	// 建立连接
	conn, err := net.Dial("tcp", mqServerAddr)
	if err != nil {
		// 失败重连
		time.Sleep(time.Second)
		goto SubscribeLoop
	}
	// 发送一个订阅消息
	subscribeMessage := Message{
		Action:        server.Subscribe,
		ConsumerGroup: "default",
		Data:          nil,
		Topic:         topicName,
	}

	msgByte, _ := json.Marshal(subscribeMessage)
	err = gotool.WriteTCPResponse(conn, msgByte)
	if err != nil {
		time.Sleep(time.Second)
		goto SubscribeLoop
	}
	// 持续监听消息的循环
	for {
		resp, err := gotool.ReadTCPResponse(conn)
		if err != nil {
			// 连接断开，尝试重连
			conn.Close()
			gotool.LogError("连接断开，尝试重连")
			break
		}
		if onMessage != nil {
			onMessage(resp)
		}
	}
	// 断线重连
	time.Sleep(time.Second)
	goto SubscribeLoop
}
