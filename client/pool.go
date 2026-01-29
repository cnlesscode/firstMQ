package client

import (
	"errors"
	"time"

	"github.com/cnlesscode/firstMQ/configs"
	"github.com/cnlesscode/gotool"
	serverFinderClient "github.com/cnlesscode/serverFinder/client"
)

// 建立连接池 :
// ServerFinderAddr  ServerFinder 服务地址,
// capacity 每个节点连接池容量,
func New(ServerFinderAddr string, capacityForNode int) (*MQPool, error) {
	// 新建连接池
	mqPool := &MQPool{
		ServerFindAddr: ServerFinderAddr,
		Addresses:      nil,
		// 总连接池 [ 缓存管道 ]
		Connections: make(chan *MQConnection, 100000),
		// 坏的连接池
		BadConnections:  make(chan *MQConnection, 100000),
		CapacityForNode: capacityForNode,
		ErrorMessage:    make(chan []byte, 1000000),
		ServerStatus:    make(map[string]bool),
	}

	// 创建监听
	go func(mqPoolIn *MQPool) {
		time.Sleep(time.Second * 5)
		serverFinderClient.Listen(
			ServerFinderAddr,
			configs.ServerFinderVarKey,
			func(message map[string]int) {
				gotool.LogDebug("Server node changes : ", message)
				if len(message) < 1 {
					return
				}
				mqPoolIn.Init(message)
			},
		)
	}(mqPool)

	// 初始化连接池
	err := mqPool.Init(nil)
	if err != nil {
		return mqPool, err
	}

	// 监听错误消息并自动发送
	RetryErrorMessages(mqPool)

	// 错误连接自动修复
	go func(mqPoolIn *MQPool) {
		for {
			select {
			case badConn := <-mqPoolIn.BadConnections:
				// 如果连接对应的服务器已经下线直接丢弃错误连接
				if !mqPoolIn.ServerStatus[badConn.Addr] {
					continue
				}
				// 有错误连接尝试修复
				conn, err := mqPoolIn.NewAConnForPool(badConn.Addr)
				if err == nil {
					// 创建成功，则将错误连接丢弃
					mqPoolIn.Connections <- conn
				} else {
					// 创建失败，则将错误连接放入错误连接池
					mqPoolIn.BadConnections <- badConn
				}
			default:
				time.Sleep(time.Second * 3)
			}
		}
	}(mqPool)

	return mqPool, err
}

func (m *MQPool) Init(addrs map[string]int) error {

	// 1. 记录旧的节点列表
	oldAddresses := m.Addresses

	// 2. 没有传递服务地址，则从服务发现中获取
	if addrs == nil {
		// 整理服务地址
		err := m.GetMQServerAddresses()
		if err != nil {
			return errors.New("无可用服务 E10001")
		}
	} else {
		m.Addresses = addrs
	}

	// 3. 检查服务器可用状态，不可用则移除
	for k := range m.Addresses {
		connIn, err := NewAConn(k)
		if err != nil {
			// 从地址map中移除
			delete(m.Addresses, k)
		} else {
			connIn.Close()
		}
	}
	addressesLen := len(m.Addresses)
	if addressesLen < 1 {
		return errors.New("无可用服务 E10002")
	}

	// 4. 如果有服务器节点掉线，发现并标注其状态
	for k := range oldAddresses {
		// 查找该节点是否在最新的服务器地址列表中
		if _, ok := m.Addresses[k]; !ok {
			m.ServerStatus[k] = false
		} else {
			m.ServerStatus[k] = true
		}
	}

	// 5. 遍历各个节点, 创建对应连接, 填充到连接池
	for addr := range m.Addresses {
		// 新的节点
		if _, ok := oldAddresses[addr]; !ok {
			m.ServerStatus[addr] = true
			for i := 0; i < m.CapacityForNode; i++ {
				tcpConnection, _ := m.NewAConnForPool(addr)
				m.Connections <- tcpConnection
			}
		}
	}

	return nil
}
