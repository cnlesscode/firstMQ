package main

import (
	"fmt"
	"runtime"
	"time"

	"github.com/cnlesscode/firstMQ/configs"
	"github.com/cnlesscode/firstMQ/server"
)

func main() {
	// 运行模式
	if configs.RunMode == "debug" || configs.RunMode == "dev" {
		go func() {
			for {
				time.Sleep(time.Second * 5)
				fmt.Printf("协程数 : %v\n", runtime.NumGoroutine())
			}
		}()
	}
	server.Start()
}
