package configs

import (
	"os"
	"path"
	"time"

	"github.com/cnlesscode/gotool"
	"github.com/cnlesscode/gotool/gfs"
	"github.com/cnlesscode/gotool/iniReader"
	"github.com/cnlesscode/serverFinder"
)

// 全局数据存储位置
var GlobalDataDir string

// 当前服务器内网IP
var CurrentIP string = gotool.GetLocalIP()

// ServerFinder 配置
var ServerFinderConfig = serverFinder.Config{
	Enable: "off",
}

// ServerFinder 服务 key
var ServerFinderVarKey string = "firstMQServers"

// FirstMQ服务配置
type FirstMQConfigStruct struct {
	// 是否开启服务
	Enable string
	// 数据存储目录名称
	DataDir string
	// 每个分片存储数据条目数
	FragmentCapacity int64
	// 主服务监听端口
	Port string
	// WebSocket 服务端口
	WebSocketPort string
	// 生产消息临时管道缓存长度
	ChannelCapactiyForProduct int
	// 落盘时每次最多读取数据数量
	MaxNumberForRepaireToDisk int64
	// 消费环节每次向管道填充消息数量
	FillNumberEachTime int64
	// 落盘时空闲休眠时间
	IdleSleepTimeForWrite time.Duration
	// 填充消费消息时空闲休眠时间
	IdleSleepTimeForFillMessage time.Duration
}

var FirstMQConfig = &FirstMQConfigStruct{}

func Init() {
	configFile := gotool.Root + "configs.ini"
	if !gfs.FileExists(configFile) {
		panic("配置文件不存在")
	}
	// 1. 初始化配置文件
	iniReader := iniReader.New(configFile)

	// 2. 全局数据目录，以系统分隔符结尾
	GlobalDataDir = path.Join(
		gotool.Root,
		iniReader.String("", "GlobalDataDirName"),
	)
	// 2.1 检查全局数据目录
	if !gfs.DirExists(GlobalDataDir) {
		err := os.Mkdir(GlobalDataDir, 0644)
		if err != nil {
			panic("数据目录创建失败: " + err.Error() + "\n")
		}
	}

	// 3. ServerFinder 服务地址
	ServerFinderConfig.Host = iniReader.String("ServerFinder", "Host")
	ServerFinderConfig.Port = iniReader.String("ServerFinder", "Port")
	ServerFinderConfig.DataLogDir = path.Join(
		GlobalDataDir,
		iniReader.String("ServerFinder", "DataLogsDirName"),
	)

	// 4. FirstMQ 服务配置
	FirstMQConfig.Enable = iniReader.String("FirstMQ", "Enable")
	FirstMQConfig.DataDir = path.Join(
		GlobalDataDir,
		iniReader.String("FirstMQ", "DataLogsDirName"))
	FirstMQConfig.FragmentCapacity = iniReader.Int64("FirstMQ", "FragmentCapacity")
	FirstMQConfig.Port = iniReader.String("FirstMQ", "Port")
	FirstMQConfig.WebSocketPort = iniReader.String("FirstMQ", "WebSocketPort")
	FirstMQConfig.ChannelCapactiyForProduct = iniReader.Int("FirstMQ", "ChannelCapactiyForProduct")
	FirstMQConfig.MaxNumberForRepaireToDisk = iniReader.Int64("FirstMQ", "MaxNumberForRepaireToDisk")
	FirstMQConfig.FillNumberEachTime =
		iniReader.Int64("FirstMQ", "FillNumberEachTime")
	FirstMQConfig.IdleSleepTimeForWrite =
		time.Duration(iniReader.Int("FirstMQ", "IdleSleepTimeForWrite")) * time.Millisecond
	FirstMQConfig.IdleSleepTimeForFillMessage =
		time.Duration(iniReader.Int("FirstMQ", "IdleSleepTimeForFillMessage")) * time.Millisecond

}
