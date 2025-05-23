package kernel

import (
	"encoding/json"
	"net"
	"time"

	"github.com/cnlesscode/firstMQ/configs"
	"github.com/cnlesscode/gotool"
	serverFinderClient "github.com/cnlesscode/serverFinder/client"
)

// 在各个集群节点上创建话题
func CreateTopicForClusters(topicName string) error {
	// 获取集群节点
	nodes, err := GetClusterNodes()
	if err != nil {
		return err
	}
	// 通知集群节点添加消费组
	for _, node := range nodes {
		conn, err := net.DialTimeout("tcp", node, time.Second*5)
		if err != nil {
			continue
		}
		defer conn.Close()
		message := SendMessageStruct{
			Action: 9,
			Topic:  topicName,
		}
		SendMessage(conn, message)
	}
	return nil
}

// 在各个集群节点上创建消费组
func CreateConsumeGroupForClusters(topicName, consumerGroup string) error {
	// 获取集群节点
	nodes, err := GetClusterNodes()
	if err != nil {
		return err
	}
	// 通知集群节点添加消费组
	for _, node := range nodes {
		conn, err := net.DialTimeout("tcp", node, time.Second*5)
		if err != nil {
			continue
		}
		defer conn.Close()
		message := SendMessageStruct{
			Action:        8,
			Topic:         topicName,
			ConsumerGroup: consumerGroup,
		}
		SendMessage(conn, message)
	}
	return nil
}

// 获取集群服务器列表
func GetClusterNodes() (map[string]string, error) {
	nodes := make(map[string]string, 0)
	res, err := serverFinderClient.Get(
		configs.ServerFinderConfig.Host+":"+configs.ServerFinderConfig.Port,
		"firstMQServers",
	)
	if err != nil {
		return nodes, err
	}
	err = json.Unmarshal([]byte(res), &nodes)
	if err != nil {
		return nodes, err
	}
	return nodes, nil
}

// 发送消息
func SendMessage(conn net.Conn, message SendMessageStruct) (ResponseMessage, error) {
	defer conn.Close()
	response := ResponseMessage{}
	msgByte, _ := json.Marshal(message)
	err := gotool.WriteTCPResponse(conn, msgByte)
	if err != nil {
		return response, err
	}
	buf, err := gotool.ReadTCPResponse(conn)
	if err != nil {
		return response, err
	}
	err = json.Unmarshal(buf, &response)
	if err != nil {
		return response, err
	}
	return response, nil
}
