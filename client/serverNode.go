package client

import (
	"encoding/json"

	"github.com/cnlesscode/firstMQ/configs"
	serverFinderClient "github.com/cnlesscode/serverFinder/client"
)

// 获取 MQ 服务列表
func (m *MQPool) GetMQServerAddresses() error {
	res, err := serverFinderClient.Get(m.ServerFindAddr, configs.ServerFinderVarKey)
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
