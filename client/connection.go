package client

import (
	"errors"
	"net"
	"time"
)

// 获取一个连接
func (m *MQConnectionPool) GetAConnection() (*MQConnection, error) {
	tryTimer := 0
	for {
		select {
		case tcpConnection := <-m.AllConnections:
			return tcpConnection, nil
		default:
			tryTimer++
			if tryTimer > 3 {
				return nil, errors.New("无法获取有效连接 E200106")
			}
		}
	}
}

// 初始化一个TCP连接
// 非连接池模式
func NewAConn(addr string) (*MQConnection, error) {
	tcpConnection := &MQConnection{
		Addr: addr,
	}
	conn, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		tcpConnection.Status = false
		return tcpConnection, err
	}
	tcpConnection.Conn = conn
	tcpConnection.Status = true
	return tcpConnection, nil
}

func (m *MQConnectionPool) NewAConnForPool(addr string) (*MQConnection, error) {
	tcpConnection, err := NewAConn(addr)
	tcpConnection.MapKey = m.Key
	return tcpConnection, err
}

// 关闭连接
func (st *MQConnection) Close() {
	st.Status = false
	st.Conn.Close()
}
