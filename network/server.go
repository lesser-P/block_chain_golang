package network

import (
	"block_chain_golang/block"
	"context"
	"crypto/rand"
	"fmt"
	log "github.com/corgi-kx/logcustom"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/multiformats/go-multiaddr"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// 在P2P网络中已发现的节点池
// key:节点ID  value:节点详细信息
// peer库是libp2p网络库的一部分，它提供了一些基本的数据结构和函数，用于处理网络中的节点（peer）。
var peerPool = make(map[string]peer.AddrInfo)
var ctx = context.Background()
var send = Send{}

// 启动本地节点
func StartNode(clier Clier) {
	// 获取本地区块最新高度
	bc := block.NewBlockchain()
	block.NewestBlockHeight = bc.GetLastBlockHeight()
	log.Infof("[*] 监听IP地址：%s 端口号：%s", ListenHost, ListenPort)
	r := rand.Reader
	// 为本地节点创建RSA密钥对
	prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		log.Panic(err)
	}
	// 创建本地节点地址信息
	sourceMultiAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%s", ListenHost, ListenPort))
	// 传入地址信息，RSA密钥对信息，生成libp2p本地host信息
	host, err := libp2p.New(ctx, libp2p.ListenAddrs(sourceMultiAddr), libp2p.Identity(prvKey))
	if err != nil {
		log.Panic(err)
	}
	// 写入全局变量本地主机信息
	localHost = host
	// 写入全局变量本地P2P节点地址详细信息
	localAddr = fmt.Sprintf("/ip4/%s/tcp/%s/p2p/%s", ListenHost, ListenPort, host.ID().Pretty())
	log.Infof("[*] 你的p2p地址信息：%s", localAddr)
	// 启动监听本地端口并传入一个处理流的函数，当本地节点接收到流的时候回调处理流的函数
	host.SetStreamHandler(protocol.ID(ProtocolID), handleStream)
	// 寻找p2p网络并加入到节点数据
	go findP2PPeer()
	// 检测节点池，如果发现网络当中节点有变动则打印到屏幕
	go monitorP2PNode()
	// 启一个go程去向其他p2p节点发送高度信息，来进行更新区块数据
	go sendVersionToPeers()
	// 启动程序的命令行输入环境
	go clier.ReceiveCMD()
	fmt.Println("本地网络节点已启动，详细信息请查看log日志")
	signalHandle()
}

// 节点退出信号处理
func signalHandle() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	send.SendSignOutToPeers()
	fmt.Println("本地节点以退出")
	time.Sleep(time.Second)
	os.Exit(0)
}

// 向其他p2p节点发送高度信息，来进行更新区块数据
func sendVersionToPeers() {
	// 如果节点池中还未存在节点的话，一直循环 直到发现以连接节点
	for {
		if len(peerPool) == 0 {
			time.Sleep(time.Second)
			continue
		} else {
			break
		}
	}
	send.SendVersionToPeers(block.NewestBlockHeight)
}

// 一个检测程序，检测当前网络中已发现的节点
func monitorP2PNode() {
	currentPeerPoolNum := 0
	for {
		peerPoolNum := len(peerPool)
		if peerPoolNum != currentPeerPoolNum && peerPoolNum != 0 {
			log.Info("===========================检测到网络中P2P节点变动，当前节点池存在节点===========================")
			for _, v := range peerPool {
				log.Info("|    ", v, "    |")
			}
			log.Info("===========================================================================================")
			currentPeerPoolNum = peerPoolNum
		} else if peerPoolNum != currentPeerPoolNum && peerPoolNum == 0 {
			log.Info("===========================检测到网络中P2P节点变动，当前节点池中没有节点===========================")
			currentPeerPoolNum = peerPoolNum
		}
		time.Sleep(time.Second)
	}
}

// 启动mdns寻找p2p网络，并等待节点连接
func findP2PPeer() {
	peerChan := initMDNS(ctx, localHost, RendezvousString)
	for {
		peer := <-peerChan
		// 把发现的节点加入节点池
		peerPool[fmt.Sprint(peer.ID)] = peer
	}
}
