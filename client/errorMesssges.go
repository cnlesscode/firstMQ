package client

import (
	"sync"
	"time"
)

func RetryErrorMessages(mqPoolIn *MQPool) {
	go func() {
		for {
			errorMessagesCount := len(mqPoolIn.ErrorMessage)
			if errorMessagesCount < 1 {
				time.Sleep(time.Second)
				continue
			}
			// 启动对应连接数数量的协程重发错误消息
			maxWorkers := mqPoolIn.CapacityForNode * len(mqPoolIn.Addresses)
			workers := errorMessagesCount
			if errorMessagesCount > maxWorkers {
				workers = maxWorkers
			}
			var wg sync.WaitGroup
			for i := 0; i < workers; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					select {
					case message := <-mqPoolIn.ErrorMessage:
						conn, err := mqPoolIn.GetAConnection()
						// 没有空闲连接
						if err != nil {
							select {
							case mqPoolIn.ErrorMessage <- message:
							default:
							}
							return
						} else {
							_, err = conn.SendBytes(message)
							// 发送错误
							if err != nil {
								select {
								case mqPoolIn.ErrorMessage <- message:
									return
								default:
									return
								}
							}
							// 将连接给回连接池
							if conn.Status {
								select {
								case mqPoolIn.Connections <- conn:
								default:
								}
							} else {
								mqPoolIn.BadConnections <- conn
							}
						}
					default:
						return
					}
				}()
			}
			wg.Wait()
			continue
		}
	}()
}
