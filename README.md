### FirstMQ
> First Message Queue ( GraceMQ ) 是一款基于 GO 语言的、完全免费的、高效的、易部署的、易使用的分布式消息中间件。

### FirstMQ 使用手册
https://www.lesscode.work/sections/94de498e71d87a0d011295bf5deb0fd2.html

### 快速启动
#### main.go
```
func main() {
	if configs.RunMode == "debug" {
		go func() {
			for {
				time.Sleep(time.Second * 5)
				fmt.Printf("协程数 : %v\n", runtime.NumGoroutine())
				fmt.Printf("cap(kernel.MessageChannels[\"test\"]): %v\n", cap(kernel.MessageChannels["test"]))
				fmt.Printf("len(kernel.MessageChannels[\"test\"]): %v\n", len(kernel.MessageChannels["test"]))
			}
		}()
	}

	log.Println("当前服务器IP:" + configs.CurrentIP)

	// 启动 ServerFinder 服务
	// 开启条件 : 服务器ip == ServerFinderConfig.Host
	go func() {
		serverFinder.Start(configs.ServerFinderConfig)
	}()

	// 开启 WS 服务
	go func() {
		server.StartWSServer()
	}()

	// 开启 TCP 服务
	server.StartFirstMQTcpServer()
}
```

#### configs.ini
```
# 运行模式 debug 调试, release 正式
RunMode=debug

# 全局数据文件存储目录名称
GlobalDataDirName=data

# ServerFinder 服务配置
[ServerFinder]
# 服务地址
Host=192.168.31.188
# 服务端口
Port=8881
# 数据文件存储目录名称
DataLogsDirName=ServerFinderDataLogs

# Message Queue 配置
[FirstMQ]
# 是否启动服务
Enable=yes
# 话题数据本地存储目录名称
DataLogsDirName=topics
# 每个分片最多存储数据数量
NumberOfFragmented=100000
# WebSocket 服务端口
WebSocketPort=8883
# 主服务端口
Port=8882
# 生产消息时，话题数据临时存储管道容量，默认100万
ChannelCapactiyForProduct=1000000
# 落盘时每次最多读取数据数量
MaxNumberForRepaireToDisk=5000
# 消费服务每次向管道填充消息数量
# 并发高时此数据设置高一些
# 最大数量不能超过一个分片存储容量
FillNumberEachTime=2000
# 落盘时空闲休眠时间，单位毫秒
# 此处延迟时间越短 CPU 占用越高
# 主机运算速度快应该设置大一些，否则循环会占用 CPU
# 经测试 50 ~ 200 毫秒比较合理
IdleSleepTimeForWrite=100
# 填充消费消息时空闲休眠时间
# 此处延迟时间越短 CPU 占用越高
# 主机运算速度快应该设置大一些，否则循环会占用 CPU
# 经测试 50 ~ 200 毫秒比较合理
IdleSleepTimeForFillMessage=200
```
