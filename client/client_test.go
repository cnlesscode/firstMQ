package client

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"
)

var addr string = "192.168.31.188:8881"

// 初始化连接池
// go test -v -run=TestInitPool
func TestInitPool(t *testing.T) {
	mqPool, err := New(addr, 100)
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("mqPool 节点服务器数量: %v\n", len(mqPool.Addresses))
	for {
		time.Sleep(time.Second * 5)
	}
}

// 创建话题
// go test -v -run=TestCreateATopic
func TestCreateATopic(t *testing.T) {
	mqPool, err := New(addr, 1)
	if err != nil {
		panic(err.Error())
	}
	// 延迟等待连接池填充
	// 生成环境无需延迟
	time.Sleep(time.Second * 1)
	// 创建话题
	response, err := mqPool.Send(Message{Action: 3, Topic: "default"})
	if err != nil {
		fmt.Printf("err: %v\n", err)
	} else {
		fmt.Printf(response.Data)
	}
}

// 生产消息 - 单条
// go test -v -run=TestProductAMessage
func TestProductAMessage(t *testing.T) {
	mqPool, err := New(addr, 1)
	if err != nil {
		panic(err.Error())
	}
	// 延迟等待连接池填充
	// 生成环境无需延迟
	time.Sleep(time.Second * 1)
	//
	response, err := mqPool.Send(Message{
		Action: 1,
		Topic:  "default",
		Data:   []byte("a test message ..."),
	})
	if err != nil {
		fmt.Printf("err: %v\n", err)
	} else {
		fmt.Printf(response.Data)
	}
}

// 生产消息 - 并发多条
// go test -v -run=TestProductMessages
func TestProductMessages(t *testing.T) {
	mqPool, err := New(addr, 100)
	if err != nil {
		panic(err.Error())
	}
	// 循环批量生产消息
	for i := 0; i < 50; i++ {
		wg := sync.WaitGroup{}
		// 开始1w个协程，并发写入
		for ii := 1; ii <= 20000; ii++ {
			n := i*10000 + ii
			wg.Add(1)
			go func(iin int) {
				defer wg.Done()
				_, err = mqPool.Send(Message{
					Action: 1,
					Topic:  "default",
					Data:   []byte(strconv.Itoa(iin) + " test message ..."),
				})
				// if err != nil {
				// 	fmt.Printf("err 0001: %v\n", err.Error())
				// }
			}(n)
		}
		wg.Wait()
		fmt.Printf("第%v次写入完成\n", i+1)
	}
	// 写入失败的消息会被记录到缓存通道中
	// 为什么会失败？
	// 客户端连接池为空，没有连接可用，这样的失败操作会被记录，客户端会自动再次提交
	// 客户端会自动重试，最终消息并不会丢失
	for {
		errCount := len(mqPool.ErrorMessage)
		fmt.Printf("errCount: %v\n", errCount)
		fmt.Printf("-- %v -- %v", len(mqPool.BadConnections), len(mqPool.Connections))
		fmt.Printf("-- %v", len(mqPool.ErrorMessage))
		time.Sleep(time.Second * 3)
	}
}

// go test -v -run=TestConsumeMessage
func TestConsumeMessage(t *testing.T) {
	mqPool, err := New(addr, 100)
	if err != nil {
		panic(err.Error())
	}
	// 启动 100个协程
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				response, _ := mqPool.Send(Message{
					Action:        2,
					Topic:         "default",
					ConsumerGroup: "default",
				})
				fmt.Printf("response: %v\n", response.Data)
			}
		}()
	}
	wg.Wait()
}

// go test -v -run=TestCreateConsumeGroup
func TestCreateConsumeGroup(t *testing.T) {
	mqPool, err := New(addr, 1)
	if err != nil {
		panic(err.Error())
	}
	// 延迟等待连接池填充
	// 生成环境无需延迟
	time.Sleep(time.Second * 1)
	//
	response, err := mqPool.Send(Message{
		Action:        7,
		Topic:         "default",
		ConsumerGroup: "consumer01",
	})
	if err != nil {
		fmt.Printf("err: %v\n", err)
	} else {
		fmt.Printf(response.Data)
	}
}

// go test -v -run=TestServerList
func TestServerList(t *testing.T) {
	mqPool, err := New(addr, 1)
	if err != nil {
		panic(err.Error())
	}
	// 延迟等待连接池填充
	// 生成环境无需延迟
	time.Sleep(time.Second * 1)
	//
	response, err := mqPool.Send(Message{
		Action: 10,
	})
	if err != nil {
		fmt.Printf("err: %v\n", err)
	} else {
		list := make(map[string]string, 0)
		err := json.Unmarshal([]byte(response.Data), &list)
		if err == nil {
			fmt.Printf("list: %v\n", list)
		} else {
			fmt.Printf("err: %v\n", err)
		}
	}
}

// go test -v -run=TestTopicList
func TestTopicList(t *testing.T) {
	mqPool, err := New(addr, 10)
	if err != nil {
		panic(err.Error())
	}
	response, err := mqPool.Send(Message{
		Action: 4,
	})
	if err != nil {
		fmt.Printf("err: %v\n", err)
	} else {
		fmt.Printf("response.Data: %v\n", response.Data)
	}
}
