package network

import "github.com/libp2p/go-libp2p-core/host"

// p2p相关，程序启动时，会被配置文件替换
var (
	RendezvousString = "meetme"       // 节点组唯一标识名称(如果节点间名称不同会找不到网络)
	ProtocolID       = "/chain/1.1.0" // 协议ID
	ListenHost       = "0.0.0.0"
	ListenPort       = "3001"
	localHost        host.Host // 本地主机
	localAddr        string    // 本地地址
)

// 交易池
var tradePool = Transactions{}

// 交易池默认大小
var TradePoolLength = 2

// 版本信息 默认0
const versionInfo = byte(0x00)

// 发送数据的头部多少位为命令
const prefixCMDLength = 12

type command string

const (
	cVersion     command = "version"
	cGetHash     command = "getHash"
	cHashMap     command = "hashMap"
	cGetBlock    command = "getBlock"
	cBlock       command = "block"
	cTransaction command = "transaction"
	cMyError     command = "myError"
)
