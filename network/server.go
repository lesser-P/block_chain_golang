package network

import (
	"context"
	"github.com/libp2p/go-libp2p-core/peer"
)

// 在P2P网络中已发现的节点池
// key:节点ID  value:节点详细信息
// peer库是libp2p网络库的一部分，它提供了一些基本的数据结构和函数，用于处理网络中的节点（peer）。
var peerPool = make(map[string]peer.AddrInfo)
var ctx = context.Background()
var send = Send{}
