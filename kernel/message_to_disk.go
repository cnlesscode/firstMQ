package kernel

import (
	"fmt"
	"os"
	"time"

	"github.com/cnlesscode/firstMQ/configs"
	"github.com/cnlesscode/gotool"
)

var _saveIndexes map[string]int64 = map[string]int64{}
var _dataLogFilePaths map[string]string = map[string]string{}
var _dataLogFileHandles map[string]*os.File = map[string]*os.File{}

// 落盘函数 : 将消息写入到临时文件
func SaveMessageToDisk(topicName string) {
	err := InitTemporaryFiles(topicName)
	if err != nil {
		gotool.LogError(err)
		return
	}
	// 初始化存储索引
	_saveIndexes[topicName], err = GetSaveIndex(topicName)
	// 启动一个协程，将临时数据写入到正式数据文件中
	go SaveTmpDataToDisk(topicName)

	// 遍历消息通道，将消息写入到临时文件中
	for {
		message := <-MessageChannels[topicName]
		SaveMessageToTemporaryFile(topicName, message)
	}
}

func SaveTmpDataToDisk(topicName string) {
	for {
		messages, _ := ReadAllMessagesFromTemporaryFile(topicName)
		if len(messages) == 0 {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		for _, message := range messages {
			SaveMessageToDiskBase(topicName, message)
		}
	}
}

func SaveMessageToDiskBase(topicName string, messages []byte) error {
	var err error
	// 根据存储索引，计算消息的起始点和结束点
	startIndex := _saveIndexes[topicName]
	// 文件索引
	fileIdx := startIndex / configs.FirstMQConfig.NumberOfFragmented
	filePath, _ := InitLogFiles(topicName, fileIdx)
	fmt.Printf("filePath: %v\n", filePath)
	// 打开文件
	// 文件曾被打开
	if _dataLogFile, ok := _dataLogFilePaths[topicName]; ok {
		if _dataLogFile == filePath {
			gotool.LogDebug("文件已经被打开")
			f, ok := _dataLogFileHandles[topicName]
			if !ok {
				f, err = os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
				if err != nil {
					return err
				}
				_dataLogFileHandles[topicName] = f
			}
			_, err := f.Write(messages)
		}
	} else {
		_dataLogFilePaths[topicName] = filePath
		_dataLogFileHandles[topicName], err = os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			return err
		}
	}
	fmt.Printf("_dataLogFilePaths: %v\n", _dataLogFilePaths)
	time.Sleep(time.Second * 1)
	return nil
}

// 落盘失败
func SaveMessageToDiskFailed(topicName string, messages [][]byte) {
	go func() {
		for _, message := range messages {
			MessageChannels[topicName] <- message
		}
	}()
}
