package server

import (
	"github.com/cnlesscode/firstMQ/configs"
	"github.com/cnlesscode/serverFinder"
)

func Start() {

	// 启动 ServerFinder 服务
	// 开启条件 : 服务器ip == ServerFinderConfig.Host
	// serverFinder 服务启动时使用了协程
	// 此处不要使用协程，保证 serverFinder 服务初始化
	serverFinder.Start(configs.ServerFinderConfig)

	// 是否开启服务
	if configs.FirstMQConfig.Enable != "yes" {
		return
	}

	// 开启 WS 服务
	go func() {
		StartWSServer()
	}()

	// 开启 TCP 服务
	StartFirstMQTcpServer()
}
