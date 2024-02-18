package network

import (
	blc "block_chain_golang/block"
	"bytes"
	"fmt"
	log "github.com/corgi-kx/logcustom"
	"sync"
	"time"
)

// 接收交易信息，满足条件后进行挖矿
func handleTransaction(content []byte) {
	t := Transactions{}
	t.Deserialize(content)
	if len(t.Ts) == 0 {
		log.Error("没有满足条件的转账信息，故不存入交易池")
	}

	// 交易池中只能存在每个地址的一笔交易内容（只能有一笔未确认的交易，这是为了防止双重支付的问题
	// 判断当前交易池中是否已有该地址发起的交易
	if len(tradePool.Ts) != 0 {
	circle:
		for i := range t.Ts {
			for _, v := range tradePool.Ts {
				if bytes.Equal(t.Ts[i].Vint[0].PublicKey, v.Vint[0].PublicKey) {
					s := fmt.Sprintf("当前交易池中，已存在此笔地址转账信息（%s），故暂不能进行转账，请等待上笔交易出块后再进行此地址转账操作", blc.GetAddressFromPublicKey(t.Ts[i].Vint[0].PublicKey))
					log.Error(s)
					// 把这笔异常的交易信息从交易列表中删除
					t.Ts = append(t.Ts[:i], t.Ts[i+1:]...)
					goto circle
				}
			}
		}
	}
	if len(t.Ts) == 0 {
		return
	}
	mineBlock(t)
}

// 调用区块模块进行挖矿操作
var lock = sync.Mutex{}

// 挖矿
func mineBlock(t Transactions) {
	// 上锁，等待上一个挖矿结束后才进行挖矿！
	lock.Lock()
	defer lock.Unlock()
	// 将临时交易池的交易添加进交易池
	tradePool.Ts = append(tradePool.Ts, t.Ts...)

	for {
		// 满足交易池规定的大小后进行挖矿
		if len(tradePool.Ts) >= TradePoolLength {
			log.Debugf("交易池已满足挖矿交易数量大小限制：%d，即将进行挖矿", TradePoolLength)
			mineTrans := Transactions{make([]Transaction, TradePoolLength)}
			copy(mineTrans.Ts, tradePool.Ts[:TradePoolLength])

			bc := blc.NewBlockchain()
			// 如果当前节点区块高度小于网络最新高度，则等待节点更新区块后再进行挖矿
			for {
				currentHeight := bc.GetLastBlockHeight()
				if currentHeight >= blc.NewestBlockHeight {
					break
				}
				time.Sleep(time.Second * 1)
			}
			// 将network下的transaction转换为blc下的transaction
			nTs := make([]blc.Transaction, len(mineTrans.Ts))
			for i := range mineTrans.Ts {
				nTs[i].TxHash = mineTrans.Ts[i].TxHash
				nTs[i].Vint = mineTrans.Ts[i].Vint
				nTs[i].Vout = mineTrans.Ts[i].Vout
			}
			// 进行转账挖矿
			bc.Transfer(nTs, send)
			// 剔除以打包进区块的交易
			newTrans := []Transaction{}
			//排除TradePoolLength个交易后的交易
			newTrans = append(newTrans, tradePool.Ts[TradePoolLength:]...)
			tradePool.Ts = newTrans
		} else {
			log.Infof("当前交易池数量：%d,交易池未满%d,暂不进行挖矿操作", len(tradePool.Ts), TradePoolLength)
			break
		}
	}
}
