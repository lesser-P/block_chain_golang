package network

import (
	"context"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery"
	"time"
)

type discoveryNotifee struct {
	PeerChan chan peer.AddrInfo
}

// 当前网络中找到新节点时，此方法会被调用
func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	n.PeerChan <- pi
}

// 启动MDNS服务
func initMDNS(ctx context.Context, peerhost host.Host, rendezvous string) chan peer.AddrInfo {
	// time.Secound检索当前网络节点的频率
	ser, err := discovery.NewMdnsService(ctx, peerhost, time.Second, rendezvous)
	if err != nil {
		panic(err)
	}

	// 注册Notifee接口类型
	n := &discoveryNotifee{}
	n.PeerChan = make(chan peer.AddrInfo)
	ser.RegisterNotifee(n)
	return n.PeerChan
}
