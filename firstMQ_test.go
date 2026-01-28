package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/cnlesscode/firstMQ/client"
)

// go test -v -run=TestSubscribe
func TestSubscribe(t *testing.T) {
	client.Subscribe("192.168.0.105:8881", "test", func(msg map[string]any) {
		fmt.Printf("msg: %v\n", msg)
	})
	for {
		time.Sleep(time.Second)
	}
}
