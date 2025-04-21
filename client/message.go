package client

import (
	"encoding/json"
	"errors"

	"github.com/cnlesscode/gotool"
)

// 发送消息
func (st *MQPool) Send(message Message) (ResponseMessage, error) {

	// 整理消息
	response := ResponseMessage{}
	messageByte, err := json.Marshal(message)
	if err != nil {
		return response, err
	}

	// 获取一个可用连接
	mqClient, err := st.GetAConnection()
	// 错误 : 无可用连接
	if err != nil {
		st.RecordErrorMessage(messageByte)
		return response, err
	}

	// 错误 : 发送消息失败
	res, err := mqClient.SendBytes(messageByte)
	if err != nil {
		st.RecordErrorMessage(messageByte)
		// 将连接给回错误连接池
		select {
		case st.BadConnections <- mqClient:
		default:
		}
		return response, err
	}

	// 至此与服务端通信时成功的，消息已经成功发送
	// 将连接给回有效连接池
	select {
	case st.Connections <- mqClient:
	default:
	}

	// 服务端返回结果错误
	err = json.Unmarshal(res, &response)
	if err != nil {
		return response, err
	}
	if response.ErrCode != 0 {
		return response, errors.New(response.Data)
	}

	// 返回结果
	return response, nil
}

// send []byte
func (st *MQConnection) SendBytes(message []byte) ([]byte, error) {

	if st.Conn == nil || !st.Status {
		return nil, errors.New("TCP 服务错误")
	}

	// ------ 发送消息 ------
	err := gotool.WriteTCPResponse(st.Conn, message)
	if err != nil {
		st.Status = false
		return nil, err
	}

	// ------ 接收消息 ------
	buf, err := gotool.ReadTCPResponse(st.Conn)
	// 连接被关闭
	if err != nil {
		st.Status = false
		return nil, err
	}

	return buf, nil
}

// 记录错误消息到缓存通道
func (st *MQPool) RecordErrorMessage(message []byte) {
	msg := Message{}
	err := json.Unmarshal(message, &msg)
	if err != nil {
		return
	}
	if msg.Action != 1 && msg.Action != 3 && msg.Action != 7 {
		return
	}
	select {
	case st.ErrorMessage <- message:
		return
	default:
		return
	}
}
