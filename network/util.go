package network

import (
	log "github.com/corgi-kx/logcustom"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"strings"
)

// 通过固定格式的地址信息，构建出P2P节点信息对象
func buildPeerInfoByAddr(addrs string) peer.AddrInfo {
	///ip4/0.0.0.0/tcp/9000/p2p/QmUyYpeMSqZp4oNMhANdG6sGeckWiGpBnzfCNvP7Pjgbvg
	// 找到首次出现/p2p/的位置+/p2p/的长度
	p2p := strings.TrimSpace(addrs[strings.Index(addrs, "/p2p/")+len("/p2p/"):])
	// 前半截
	ipTcp := addrs[:strings.Index(addrs, "/p2p/")]

	// 通过ip与端口获得multiAddr，用于处理多地址，多地址是一种自我描述的网络地址格式，他可以包含多种协议的信息，例如ip地址，端口号，协议类型等
	multiAddr, err := multiaddr.NewMultiaddr(ipTcp)
	if err != nil {
		log.Debug(err)
	}
	m := []multiaddr.Multiaddr{multiAddr}
	// 获得host.ID
	id, err := peer.IDB58Decode(p2p)
	if err != nil {
		log.Error(err)
	}
	return peer.AddrInfo{ID: id, Addrs: m}
}

// 默认前十二位为命令名称
func jointMessage(cmd command, content []byte) []byte {
	b := make([]byte, prefixCMDLength)
	for i, v := range []byte(cmd) {
		b[i] = v
	}
	joint := make([]byte, 0)
	joint = append(b, content...)
	return joint
}

// 默认前十二位为命令名称
func splitMessage(message []byte) (cmd string, content []byte) {
	cmdBytes := message[:prefixCMDLength]
	newCMDBytes := make([]byte, 0)
	for _, v := range cmdBytes {
		if v != byte(0) {
			newCMDBytes = append(newCMDBytes, v)
		}
	}
	cmd = string(newCMDBytes)
	content = message[prefixCMDLength:]
	return
}
