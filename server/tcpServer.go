package server

import (
	"net"

	"github.com/cnlesscode/firstMQ/configs"
	"github.com/cnlesscode/firstMQ/kernel"
	"github.com/cnlesscode/gotool"

	serverFinderClient "github.com/cnlesscode/serverFinder/client"
)

// 创建TCP服务器
func NewTCPServer(addr string) *TCPServer {
	// 创建 Socket 端口监听
	// listener 是一个用于面向流的网络协议的公用网络监听器接口，
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}
	// 返回实例
	return &TCPServer{listener: listener}
}

// Accept 等待客户端连接
func (t *TCPServer) Accept() {
	// 处理客户端连接
	// 关闭接口解除阻塞的 Accept 操作并返回错误
	defer t.listener.Close()
	// 循环等待客户端连接
	for {
		// 等待客户端连接
		conn, err := t.listener.Accept()
		if err == nil {
			go t.Handle(conn)
		}
	}
}

// Handle 处理客户端连接
func (t *TCPServer) Handle(conn net.Conn) {
	for {
		content, err := gotool.ReadTCPResponse(conn)
		if err != nil {
			conn.Close()
			break
		}

		// 解析消息
		message, messageByte := TCPResponse(content)
		// 订阅事件
		if message.Action == Subscribe {
			// 提取纯 IP（不含端口）
			remoteTCPAddr, ok := conn.RemoteAddr().(*net.TCPAddr)
			if !ok {
				conn.Close()
				break
			}
			clientIP := remoteTCPAddr.IP.String()

			subscribeClientsMutex.Lock()
			if _, exist := subscribeClients[message.Topic]; !exist {
				subscribeClients[message.Topic] = make(map[string]chan *SubscribeClient)
			}
			if _, exist := subscribeClients[message.Topic][clientIP]; !exist {
				subscribeClients[message.Topic][clientIP] = make(chan *SubscribeClient, 10000)
			}
			subscribeClientsMutex.Unlock()
			subscribeClients[message.Topic][clientIP] <- NewASubscribeClient(
				conn,
				message.Topic,
				clientIP,
			)

		} else {
			// 输出响应
			err = gotool.WriteTCPResponse(conn, messageByte)
			if err != nil {
				conn.Close()
				break
			}
		}
	}
}

// 开启 TCP 服务
func StartFirstMQTcpServer() {
	// 1. 注册服务到 ServerFinder
	serverFinderClient.Regist(
		configs.ServerFinderConfig.Host+":"+configs.ServerFinderConfig.Port,
		configs.ServerFinderVarKey,
		configs.CurrentIP+":"+configs.FirstMQConfig.Port,
		nil,
	)
	// 2. 初始化 FirstMQ 话题
	kernel.LoadTopics()
	// 3. 启动 FirstMQ TCP 服务
	tcpServer := NewTCPServer(":" + configs.FirstMQConfig.Port)
	gotool.LogOk(
		"FirstMQ : MQSetver is running on port ",
		configs.FirstMQConfig.Port, ".")
	tcpServer.Accept()
}
