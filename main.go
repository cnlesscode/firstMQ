package main

import (
	"time"

	"github.com/cnlesscode/firstMQ/server"
)

func main() {
	go func() {
		for {
			time.Sleep(time.Second * 5)
			// fmt.Printf("协程数 : %v\n", runtime.NumGoroutine())
		}
	}()
	server.Start()
}
