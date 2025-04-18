package server

import (
	"github.com/cnlesscode/firstMQ/configs"
	"github.com/cnlesscode/serverFinder"
)

func Start() {
	// 是否开启服务
	if configs.FirstMQConfig.Enable != "yes" {
		return
	}

	// 启动 ServerFinder 服务
	// 开启条件 : 服务器ip == ServerFinderConfig.Host
	go func() {
		serverFinder.Start(configs.ServerFinderConfig)
	}()

	// 开启 WS 服务
	go func() {
		StartWSServer()
	}()

	// 开启 TCP 服务
	StartFirstMQTcpServer()
}
