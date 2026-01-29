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
func Subscribe(ServerFinderAddr, topicName string, onMessage func(msg []byte)) {
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
		go subscribeBase(k, topicName, onMessage)
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
		Data:          []byte(topicName),
		Topic:         topicName,
	}
	msgByte, _ := json.Marshal(subscribeMessage)
	err = gotool.WriteTCPResponse(conn, msgByte)
	if err != nil {
		time.Sleep(time.Second)
		goto SubscribeLoop
	}
	gotool.LogOk("ok")
	// 持续监听消息的循环
	for {
		resp, err := gotool.ReadTCPResponse(conn)
		if err != nil {
			// 连接断开，尝试重连
			conn.Close()
			gotool.LogError("连接断开，尝试重连")
			break
		}
		if err != nil {
			continue // 跳过错误消息，继续监听
		}
		if onMessage != nil {
			onMessage(resp)
		}
	}
	gotool.LogError(".....")
	// 断线重连
	time.Sleep(time.Second)
	goto SubscribeLoop
}
