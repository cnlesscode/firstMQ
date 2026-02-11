package kernel

import (
	"time"

	"github.com/cnlesscode/firstMQ/configs"
	"github.com/cnlesscode/gotool"
)

// 落盘函数
func SaveMessageToDisk(topicName string) {
	var err error
	saveST, err := NewASaveMessageToDiskStruct(topicName)
	if err != nil {
		gotool.LogFatal("Save Message To Disk Failed! Error : ", err.Error())
		return
	}
SaveMessageToDiskStart:
	// 遍历话题内消息
	var messageCount int64 = 0
	// 从管道内取出消息
	messageSlice := make([][]byte, 0)
	for {
		getedValue := false
		select {
		case message := <-MessageChannels[topicName]:
			messageSlice = append(messageSlice, message)
			getedValue = true
		default:
			getedValue = false
		}
		if getedValue {
			messageCount++
			if messageCount >= configs.FirstMQConfig.MaxNumberForRepaireToDisk {
				break
			}
		} else {
			if messageCount > 0 {
				break
			} else {
				// 此处延迟时间越短 CPU 占用越高
				// 主机运算速度快应该设置大一些，否则循环会占用 CPU
				// 经测试 50 ~ 100 毫秒比较合理
				time.Sleep(configs.FirstMQConfig.IdleSleepTimeForWrite)
			}
		}
	}
	// 落盘
	successCount, err := saveST.SaveLogsToFile(topicName, messageSlice)
	// 如果落盘失败，回滚消息
	if err != nil {
		if successCount > 0 {
			SaveMessageToDiskFailed(topicName, messageSlice[successCount:])
		} else {
			SaveMessageToDiskFailed(topicName, messageSlice)
		}
		goto SaveMessageToDiskStart
	}

	// 成功 : 继续落盘
	goto SaveMessageToDiskStart
}

// 落盘失败
func SaveMessageToDiskFailed(topicName string, messages [][]byte) {
	go func() {
		for _, message := range messages {
			MessageChannels[topicName] <- message
		}
	}()
}
