package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	serverFinderClient "github.com/cnlesscode/serverFinder/client"
)

// 已有连接池 map
// 利用 map 键实现单利
var MQPoolMap map[string]*MQConnectionPool = make(map[string]*MQConnectionPool)

// 建立连接池 :
// ServerFinderAddr  ServerFinder 服务地址,
// capacity 连接池容量,
// purpose 用途，由于单例区分
func New(ServerFinderAddr string, capacity int, purpose string) (*MQConnectionPool, error) {
	mapKey := ServerFinderAddr + purpose
	_, ok := MQPoolMap[mapKey]
	// 已有连接池直接返回
	if ok {
		return MQPoolMap[mapKey], nil
	}
	// 新建连接池
	MQPoolMap[mapKey] = &MQConnectionPool{
		Key:            mapKey,
		ServerFindAddr: ServerFinderAddr,
		Addresses:      nil,
		AddressesLen:   0,
		// 总连接池 [ 缓存管道 ]
		AllConnections: make(chan *MQConnection, capacity+1000),
		// 对应服务器节点的连接池
		MQConnections:        make(map[string]chan *MQConnection),
		Capacity:             capacity,
		ConnNumber:           make(map[string]int),
		ConnDifferenceNumber: make(map[string]int),
		ErrorMessage:         make(chan []byte, 1000000),
		ServerStatus:         make(map[string]bool),
	}

	// 创建监听
	go func() {
		time.Sleep(time.Second * 5)
		serverFinderClient.Listen(
			ServerFinderAddr,
			"firstMQServers",
			func(message map[string]any) {
				fmt.Printf("message: %v\n", message)
				if len(message) < 1 {
					return
				}
				MQPoolMap[mapKey].Init(message)
			},
		)
	}()

	// 初始化连接池
	err := MQPoolMap[mapKey].Init(nil)
	if err != nil {
		return MQPoolMap[mapKey], err
	}

	// 监听错误消息并自动发送
	go func(mapKeyIn string) {
		for {
			errorMessagesCount := len(MQPoolMap[mapKeyIn].ErrorMessage)
			if errorMessagesCount < 1 {
				time.Sleep(time.Second)
				continue
			}
			// 启动对应连接数数量的协程重发错误消息
			var wg sync.WaitGroup
			for i := 0; i < MQPoolMap[mapKeyIn].Capacity*10; i++ {
				wg.Add(1)
				go func(mapKeyIn string) {
					defer wg.Done()
					select {
					case message := <-MQPoolMap[mapKeyIn].ErrorMessage:
						conn, err := MQPoolMap[mapKeyIn].GetAConnection()
						// 如果有错再放回错误连消息chi
						if err != nil {
							select {
							case MQPoolMap[mapKeyIn].ErrorMessage <- message:
								return
							default:
								return
							}
						} else {
							_, err = conn.SendBytes(message)
							if err != nil {
								select {
								case MQPoolMap[mapKeyIn].ErrorMessage <- message:
									return
								default:
									return
								}
							}
						}
					default:
						return
					}
				}(mapKeyIn)
			}
			wg.Wait()
			continue
		}
	}(mapKey)

	return MQPoolMap[mapKey], err
}

func (m *MQConnectionPool) Init(addrs map[string]any) error {

	// 1. 没有传递服务地址，则从服务发现中获取
	if addrs == nil {
		// 整理服务地址
		err := m.GetMQServerAddresses()
		if err != nil {
			return errors.New("无可用服务 E10001")
		}
	} else {
		m.Addresses = addrs
	}
	// 2. 检查服务器可用状态，不可用则移除
	for k, v := range m.Addresses {
		connIn, err := NewAConn(v.(string))
		if err != nil {
			// 从地址map中移除
			delete(m.Addresses, k)
			// 从 ServerFinder 中移除
			serverFinderClient.Remove(m.ServerFindAddr, "firstMQServers", k)
		} else {
			connIn.Close()
		}
	}
	addressesLen := len(m.Addresses)
	if addressesLen < 1 {
		return errors.New("无可用服务 E10002")
	}

	// 3. 如果有服务器节点掉线，发现并标注其状态
	for k := range m.MQConnections {
		// 查找该节点是否在最新的服务器地址列表中
		if _, ok := m.Addresses[k]; !ok {
			m.ServerStatus[k] = false
			m.ConnNumber[k] = 0
		} else {
			m.ServerStatus[k] = true
		}
	}

	// 4. 计算每个节点应该建立的连接数
	m.AddressesLen = addressesLen
	if m.Capacity < m.AddressesLen {
		m.Capacity = m.AddressesLen
	}
	capacityForEveryServer := m.Capacity / m.AddressesLen

	// 5. 遍历各个节点, 创建对应连接池
	for addr := range m.Addresses {
		// 新的节点
		if _, ok := m.MQConnections[addr]; !ok {
			m.ConnNumber[addr] = capacityForEveryServer
			m.ServerStatus[addr] = true
			m.InitNewNode(addr)
		} else {
			// 已有节点
			m.ConnDifferenceNumber[addr] = capacityForEveryServer - m.ConnNumber[addr]
			m.ConnNumber[addr] = capacityForEveryServer
			if m.ConnDifferenceNumber[addr] > 0 {
				for i := 0; i < m.ConnDifferenceNumber[addr]; i++ {
					tcpConnection, _ := m.NewAConnForPool(addr)
					m.MQConnections[addr] <- tcpConnection
				}
				log.Println("✔ 新增连接 : ", addr, " 完成，新增数量:", m.ConnDifferenceNumber[addr])
				m.ConnDifferenceNumber[addr] = 0
			} else if m.ConnDifferenceNumber[addr] < 0 {
				go func() {
					for i := 0; i > m.ConnDifferenceNumber[addr]; i-- {
						tcpConnection := <-m.AllConnections
						if tcpConnection.Addr == addr {
							if tcpConnection.Conn != nil {
								tcpConnection.Conn.Close()
							}
						} else {
							m.AllConnections <- tcpConnection
						}
					}
					m.ConnDifferenceNumber[addr] = 0
					log.Println("✔ 减少连接 : ", addr, " 完成，减少数量:", m.ConnDifferenceNumber[addr])
				}()
			}
		}
	}
	return nil
}

// 初始化新节点服务器的连接
func (m *MQConnectionPool) InitNewNode(addr string) {
	m.MQConnections[addr] = make(chan *MQConnection, m.ConnNumber[addr]+10)
	m.ConnDifferenceNumber[addr] = 0
	for i := 0; i < m.ConnNumber[addr]; i++ {
		tcpConnection, err := m.NewAConnForPool(addr)
		if err != nil {
			continue
		}
		m.MQConnections[addr] <- tcpConnection
	}

	// 从节点连接池中获取连接，填充进总连接池
	go func(addrIn string) {
		for {
			select {
			case tcpConnection := <-m.MQConnections[addrIn]:
				if tcpConnection.Status {
					m.AllConnections <- tcpConnection
				}
			default:
				time.Sleep(time.Millisecond * 10)
			}
		}
	}(addr)
}

// 获取 MQ 服务列表
func (m *MQConnectionPool) GetMQServerAddresses() error {
	res, err := serverFinderClient.Get(m.ServerFindAddr, "firstMQServers")
	if err != nil {
		return err
	}
	address := make(map[string]any)
	err = json.Unmarshal([]byte(res), &address)
	if err != nil {
		return err
	}
	m.Addresses = address
	return nil
}
