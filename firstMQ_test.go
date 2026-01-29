package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/cnlesscode/firstMQ/client"
)

func subscribeOnMessage(message []byte) {
	fmt.Printf("Received message : %s", message)
}

// go test -v -run=TestSubscribe
func TestSubscribe(t *testing.T) {
	// 注意 : Subscribe是异步执行的
	client.Subscribe("192.168.0.105:8881", "default", subscribeOnMessage)
	for {
		time.Sleep(time.Second)
	}
}
