package main

import (
	"fmt"
	"runtime"
	"time"

	"github.com/cnlesscode/firstMQ/configs"
	"github.com/cnlesscode/firstMQ/kernel"
	"github.com/cnlesscode/firstMQ/server"
	"github.com/cnlesscode/serverFinder"
)

func main() {

	// 是否开启服务
	if configs.FirstMQConfig.Enable != "yes" {
		return
	}

	// 运行模式
	if configs.RunMode == "debug" || configs.RunMode == "dev" {
		go func() {
			for {
				time.Sleep(time.Second * 5)
				fmt.Printf("协程数 : %v\n", runtime.NumGoroutine())
				fmt.Printf("cap(kernel.MessageChannels[\"test\"]): %v\n", cap(kernel.MessageChannels["test"]))
				fmt.Printf("len(kernel.MessageChannels[\"test\"]): %v\n", len(kernel.MessageChannels["test"]))
			}
		}()
	}

	// 启动 ServerFinder 服务
	// 开启条件 : 服务器ip == ServerFinderConfig.Host
	go func() {
		serverFinder.Start(configs.ServerFinderConfig)
	}()

	// 开启 WS 服务
	go func() {
		server.StartWSServer()
	}()

	// 开启 TCP 服务
	server.StartFirstMQTcpServer()
}
