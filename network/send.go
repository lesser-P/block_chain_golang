package network

import (
	"block_chain_golang/block"
	"bufio"
	log "github.com/corgi-kx/logcustom"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
)

type Send struct {
}

// 向网络中其他节点发送本节点退出信号
func (s Send) SendSignOutToPeers() {
	ss := "节点" + localAddr + "已退出网络"
	m := myerror{ss, localAddr}
	data := jointMessage(cMyError, m.serialize())
	for _, v := range peerPool {
		s.SendMessage(v, data)
	}
}

// 向网络中其他节点发送高度信息
func (s Send) SendVersionToPeers(lastHeight int) {
	newV := version{Version: versionInfo, Height: lastHeight, AddrFrom: localAddr}
	data := jointMessage(cVersion, newV.serialize())
	for _, v := range peerPool {
		s.SendMessage(v, data)
	}
}

// 向网络中其他节点发送交易信息
func (s Send) SendTransToPeers(ts []block.Transaction) {
	// 向交易信息列表加入节点地址信息
	nts := make([]Transaction, len(ts))
	for i := range ts {
		nts[i].TxHash = ts[i].TxHash
		nts[i].Vout = ts[i].Vout
		nts[i].Vint = ts[i].Vint
		// 把所有交易的来源地址变成本地地址
		nts[i].AddrFrom = localAddr
	}
	tss := Transactions{nts}
	// 开启一个go程,先传送给自己进行处理
	go handleTransaction(tss.Serialize())
	// 然后将命令与交易列表拼接好发送给全网节点
	data := jointMessage(cTransaction, tss.Serialize())
	log.Tracef("准备发送%d笔交易到网络中其他p2p节点", len(tss.Ts))
	for _, v := range peerPool {
		s.SendMessage(v, data)
	}
}

// 基础发送信息方法
func (s Send) SendMessage(peer peer.AddrInfo, data []byte) {
	// 连接传入的对等节点，建立与一个对等节点peer的连接
	if err := localHost.Connect(ctx, peer); err != nil {
		log.Error("Connection failed", err)
	}
	// 打开一个流，向流写入信息后关闭
	stream, err := localHost.NewStream(ctx, peer.ID, protocol.ID(ProtocolID))
	if err != nil {
		log.Debug("Stream open failed", err)
	} else {
		cmd, _ := splitMessage(data)
		// 创建一个缓冲流的容器，新的读写器
		rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
		// 写入信息到缓冲容器中
		_, err = rw.Write(data)
		if err != nil {
			log.Panic(err)
		}
		// 向流中写入所有缓冲数据
		err := rw.Flush()
		if err != nil {
			log.Panic(err)
		}
		// 关闭流
		err = stream.Close()
		if err != nil {
			log.Panic(err)
		}
		// 数据会被发送到指定的peer.ID节点中
		log.Debugf("send cmd : %s to peer:%v", cmd, peer)
	}
}
