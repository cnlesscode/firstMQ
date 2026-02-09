package kernel

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"sync"

	"github.com/cnlesscode/firstMQ/configs"
)

var _temporaryDataFilePaths map[string]string = make(map[string]string, 0)
var _temporaryDataFileHandles map[string]*os.File = make(map[string]*os.File, 0)
var _temporaryFileLockers map[string]*sync.Mutex = make(map[string]*sync.Mutex, 0)

func InitTemporaryFiles(topicName string) error {
	var err error = nil

	// 临时数据文件
	_temporaryDataFilePaths[topicName] = path.Join(
		configs.FirstMQConfig.DataDir,
		topicName,
		"temporary_data.bin",
	)
	_temporaryDataFileHandles[topicName], err = os.OpenFile(
		_temporaryDataFilePaths[topicName],
		os.O_CREATE|os.O_RDWR|os.O_APPEND,
		0644)
	if err != nil {
		return err
	}

	// 定义文件锁
	_temporaryFileLockers[topicName] = &sync.Mutex{}

	return nil
}

func SaveMessageToTemporaryFile(topicName string, message []byte) error {

	// 检查文件锁
	if _, ok := _temporaryFileLockers[topicName]; !ok {
		return errors.New("临时文件对应锁不存在")
	}
	_temporaryFileLockers[topicName].Lock()
	defer _temporaryFileLockers[topicName].Unlock()

	// 将数据保存的临时文件
	if _, ok := _temporaryDataFileHandles[topicName]; !ok {
		return errors.New("临时文件不存在")
	}

	// 1. 写入消息长度 [ 8字节，小端序 ]
	lengthBuf := make([]byte, 8)
	binary.LittleEndian.PutUint64(lengthBuf, uint64(len(message)))
	if _, err := _temporaryDataFileHandles[topicName].Write(lengthBuf); err != nil {
		_temporaryDataFileHandles[topicName].Close()
		return err
	}

	// 2. 写入消息内容
	if _, err := _temporaryDataFileHandles[topicName].Write(message); err != nil {
		_temporaryDataFileHandles[topicName].Close()
		return err
	}

	return nil
}

func ReadAllMessagesFromTemporaryFile(topicName string) ([][]byte, error) {
	// 检查文件锁
	if _, ok := _temporaryFileLockers[topicName]; !ok {
		return nil, errors.New("临时文件对应锁不存在")
	}
	_temporaryFileLockers[topicName].Lock()
	defer _temporaryFileLockers[topicName].Unlock()

	if _, ok := _temporaryDataFileHandles[topicName]; !ok {
		return nil, errors.New("临时文件不存在")
	}

	file := _temporaryDataFileHandles[topicName]
	file.Seek(0, io.SeekStart)
	var messages [][]byte

	// ## 注意 read 和 readFull 都会自动移动文件指针！
	for {

		// 1. 读取消息长度（8字节）
		lengthBuf := make([]byte, 8)
		_, err := file.Read(lengthBuf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break // 正常结束
			}
			return nil, fmt.Errorf("read message length: %w", err)
		}
		length := binary.LittleEndian.Uint64(lengthBuf)

		// 2. 读取指定长度的消息
		message := make([]byte, length)
		// 使用 ReadFull 确保读满
		_, err = io.ReadFull(_temporaryDataFileHandles[topicName], message)
		if err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}
	return messages, nil
}
