package client

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cnlesscode/firstMQ/configs"
	"github.com/cnlesscode/gotool"
	serverFinderClient "github.com/cnlesscode/serverFinder/client"
	"github.com/gorilla/websocket"
)

var subscribeConnections map[string]*MQConnection = make(map[string]*MQConnection)

// 客户端订阅话题
func Subscribe(ServerFinderAddr, topicName string, onMessage func(msg map[string]any)) {
	// 获取集群服务器地址
	res, err := serverFinderClient.Get(ServerFinderAddr, configs.ServerFinderVarKey)
	if err != nil {
		return
	}
	addresses := make(map[string]any)
	err = json.Unmarshal([]byte(res), &addresses)
	if err != nil {
		return
	}
	for k := range addresses {
		subscribeBase(k, topicName)
	}

	// 监听服务器组
	time.Sleep(time.Second * 10)
	serverFinderClient.Listen(
		ServerFinderAddr,
		configs.ServerFinderVarKey,
		func(message map[string]any) {
			gotool.LogDebug("Server node changes : ", message)
			if len(message) < 1 {
				return
			}
		},
	)
}

func subscribeBase(mqServerAddr, topicName string) {
	var keyName = mqServerAddr + "_" + topicName
	fmt.Printf("keyName: %v\n", keyName)
	if _, ok := subscribeConnections[keyName]; ok {
		return
	}
	go func(_mqServerAddr, _topicName string) {
		// 初始化连接地址
		url := "ws://" + mqServerAddr
	SubscribeLoop:
		// 建立连接
		conn, _, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			// 失败重连
			time.Sleep(time.Second)
			goto SubscribeLoop
		}
		// 监听消息
		for {
			_, messageByte, err := conn.ReadMessage()
			if err != nil {
				// 断开连接
				conn.Close()
				break
			}
			fmt.Printf("messageByte: %s\n", messageByte)
		}
		// 断线重连
		time.Sleep(time.Second)
		goto SubscribeLoop
	}(mqServerAddr, topicName)
}
