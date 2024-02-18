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
	"github.com/multiformats/go-multiaddr"
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

}
