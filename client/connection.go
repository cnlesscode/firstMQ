package client

import (
	"errors"
	"net"
	"time"
)

// 获取一个连接
func (m *MQPool) GetAConnection() (*MQConnection, error) {
	select {
	case tcpConnection := <-m.Connections:
		return tcpConnection, nil
	case <-time.After(time.Second):
		return nil, errors.New("无法获取有效连接 E200103")
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

func (m *MQPool) NewAConnForPool(addr string) (*MQConnection, error) {
	tcpConnection, err := NewAConn(addr)
	return tcpConnection, err
}

// 关闭连接
func (st *MQConnection) Close() {
	st.Status = false
	st.Conn.Close()
}
