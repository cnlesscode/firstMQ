package kernel

import (
	"encoding/binary"
	"os"
	"path/filepath"

	"github.com/cnlesscode/firstMQ/configs"
)

type SaveMessageToDiskStruct struct {
	TopicName string

	MaxSaveIndex           int64
	MaxSaveIndexFilePath   string
	MaxSaveIndexFileHandle *os.File

	DataLogFilePath             string
	DataLogFileHandle           *os.File
	DataLogFileWriteStartOffset int64

	DataIndexFilePath             string
	DataIndexFileHandle           *os.File
	DataIndexFileWriteStartOffset int64
}

func NewASaveMessageToDiskStruct(topicName string) (*SaveMessageToDiskStruct, error) {
	st := &SaveMessageToDiskStruct{
		TopicName: topicName,
	}
	var err error

	// 初始化存储索引文件
	st.MaxSaveIndexFilePath = filepath.Join(
		configs.FirstMQConfig.DataDir,
		topicName,
		"save_index.bin",
	)
	st.MaxSaveIndexFileHandle, err = os.OpenFile(
		st.MaxSaveIndexFilePath,
		os.O_CREATE|os.O_RDWR,
		0644,
	)
	if err != nil {
		return st, err
	}
	// 获取存储索引值
	st.GetSaveIndex()
	return st, nil
}

func (s *SaveMessageToDiskStruct) GetSaveIndex() error {
	var saveIndex int64 = 0
	err := binary.Read(s.MaxSaveIndexFileHandle, binary.LittleEndian, &saveIndex)
	if err != nil {
		return err
	}
	s.MaxSaveIndex = saveIndex
	return nil
}

func (s *SaveMessageToDiskStruct) SaveLogsToFile(topicName string, messages [][]byte) (int64, error) {
	var err error
	var successCount int64 = 0
	err = s.GetSaveIndex()
	// 计算存储文件索引值
	fileIndex := s.MaxSaveIndex / configs.FirstMQConfig.FragmentCapacity
	// 初始化日志文件路径和索引文件路径
	dataLogFilePath, dataIndexFilePath := InitLogFiles(s.TopicName, fileIndex)
	// 如果数据文件路径不一致，则关闭文件句柄
	if dataLogFilePath != s.DataLogFilePath {
		s.DataLogFilePath = dataLogFilePath
		if s.DataLogFileHandle != nil {
			s.DataLogFileHandle.Close()
		}
	}
	// 如果索引文件路径不一致，则关闭文件句柄
	if dataIndexFilePath != s.DataIndexFilePath {
		s.DataIndexFilePath = dataIndexFilePath
		if s.DataIndexFileHandle != nil {
			s.DataIndexFileHandle.Close()
		}
	}
	// 初始化日志文件句柄
	s.DataLogFileHandle, err = os.OpenFile(
		s.DataLogFilePath,
		os.O_CREATE|os.O_RDWR|os.O_APPEND,
		0644)
	if err != nil {
		return 0, err
	}
	// 初始化索引文件句柄
	s.DataIndexFileHandle, err = os.OpenFile(
		s.DataIndexFilePath,
		os.O_CREATE|os.O_RDWR|os.O_APPEND,
		0644,
	)
	if err != nil {
		return 0, err
	}

	// 检查消息总量是否超过最大存储数量
	// 计算消息总数
	var messageLength int64 = int64(len(messages))
	dataLogStartIndexForCurrentFile := s.MaxSaveIndex % configs.FirstMQConfig.FragmentCapacity
	// 消息会跨越两个分片
	if dataLogStartIndexForCurrentFile+messageLength > configs.FirstMQConfig.FragmentCapacity {
		// 截断消息
		/*
			| 起始值 | 分片容量  | 消息总数 | 截取长度 |
			|   0    |   5     |    3   |     3  |
			|   3    |   5     |    3   |     2  |
		*/
		currentFileMessageCount := configs.FirstMQConfig.FragmentCapacity - dataLogStartIndexForCurrentFile
		if dataLogStartIndexForCurrentFile > 0 {
			fileInfo, err := s.DataLogFileHandle.Stat()
			if err != nil {
				return 0, err
			}
			s.DataLogFileWriteStartOffset = fileInfo.Size()
		}
		messagesForCruuentFile := messages[0:currentFileMessageCount]
		err = s.SaveDataLogsBase(messagesForCruuentFile)
		if err != nil {
			return 0, err
		}
		successCount = currentFileMessageCount
		// 存储剩余消息
		// 剩余消息将存储到新的分片
		// 先关闭当前分片文件句柄
		s.DataIndexFileHandle.Close()
		s.DataLogFileHandle.Close()
		// 新建一个分片存储剩余数据
		s.DataLogFilePath, s.DataIndexFilePath = InitLogFiles(s.TopicName, fileIndex+1)
		s.DataLogFileHandle, err = os.OpenFile(
			s.DataLogFilePath,
			os.O_CREATE|os.O_RDWR|os.O_APPEND,
			0644,
		)
		if err != nil {
			return successCount, err
		}
		// 新建一个索引文件
		s.DataIndexFileHandle, err = os.OpenFile(
			s.DataIndexFilePath,
			os.O_CREATE|os.O_RDWR|os.O_APPEND,
			0644,
		)
		if err != nil {
			return successCount, err
		}
		s.DataLogFileWriteStartOffset = 0
		s.DataIndexFileWriteStartOffset = 0
		err = s.SaveDataLogsBase(messages[currentFileMessageCount:])
		if err != nil {
			return successCount, err
		}
	} else { // 消息不会跨越两个分片
		if dataLogStartIndexForCurrentFile > 0 {
			fileInfo, err := s.DataLogFileHandle.Stat()
			if err != nil {
				return 0, err
			}
			s.DataLogFileWriteStartOffset = fileInfo.Size()
		}
		err = s.SaveDataLogsBase(messages)
		if err != nil {
			return successCount, err
		}
	}
	return messageLength, nil
}

func (s *SaveMessageToDiskStruct) SaveDataLogsBase(messages [][]byte) error {
	originalDataLogFileWriteStartOffset := s.DataLogFileWriteStartOffset
	originalDataIndexFileWriteStartOffset := s.DataIndexFileWriteStartOffset
	indexData := make([]int64, 0)
	dataPrepare := make([]byte, 0)
	messageCount := int64(len(messages))
	for _, msg := range messages {
		dataPrepare = append(dataPrepare, msg...)
		indexData = append(indexData, s.DataLogFileWriteStartOffset)
		dataLen := int64(len(msg))
		indexData = append(indexData, dataLen)
		s.DataLogFileWriteStartOffset += dataLen
		s.DataIndexFileWriteStartOffset += 16
	}
	// 1. 存储消息数据
	_, err := s.DataLogFileHandle.Write(dataPrepare)
	if err != nil {
		return err
	}
	// 2. 存储索引数据
	err = binary.Write(s.DataIndexFileHandle, binary.LittleEndian, indexData)
	if err != nil {
		// 如果出错需要回退日志数据
		s.DataLogFileHandle.Truncate(originalDataLogFileWriteStartOffset)
		s.DataLogFileWriteStartOffset = originalDataLogFileWriteStartOffset
		return err
	}
	// 3. 记录 saveIndex
	s.MaxSaveIndex += messageCount
	s.MaxSaveIndexFileHandle.Seek(0, 0)
	err = binary.Write(
		s.MaxSaveIndexFileHandle,
		binary.LittleEndian,
		s.MaxSaveIndex)
	s.MaxSaveIndexFileHandle.Truncate(8)
	if err != nil {
		// 3.3.1 如果出错需要回退索日志数据
		s.DataLogFileHandle.Truncate(originalDataLogFileWriteStartOffset)
		s.DataIndexFileWriteStartOffset = originalDataLogFileWriteStartOffset
		// 3.3.2 如果出错需要回退索引数据
		s.DataIndexFileHandle.Truncate(originalDataIndexFileWriteStartOffset)
		s.DataIndexFileWriteStartOffset = originalDataIndexFileWriteStartOffset
		return err
	}

	s.MaxSaveIndexFileHandle.Sync()
	s.DataLogFileHandle.Sync()
	s.DataIndexFileHandle.Sync()

	return nil
}
