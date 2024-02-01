package block

import (
	"block_chain_golang/database"
	"fmt"
	log "github.com/corgi-kx/logcustom"
)

type blockchain struct {
	DB *database.BlockchainDB //封装blot结构体
}

func NewBlockchain() *blockchain {
	return &blockchain{database.New()}
}

// 创建创世区块交易信息
func (bc *blockchain) CreateGenesisTransaction(address string, value int, send Sender) {
	// 判断地址格式是否正确
	if !IsVailBitcoinAddress(address) {
		log.Error("地址错误", address)
		return
	}

	//创世区块数据
	txi := TxInput{
		[]byte{},
		-1,
		nil,
		nil,
	}

	// 本地存创世区块的公私钥信息
	wallets := NewWallets(bc.DB)
	//创世区块地址的公私钥信息
	genesisKeys, ok := wallets.Wallets[address]
	if !ok {
		log.Error("没有找到对应地址的公私钥信息")
	}
	publicKeyHash := generatePublicKeyHash(genesisKeys.PublicKey)
	txo := TxOutput{Value: value, PublicKeyHash: publicKeyHash}
	ts := Transaction{nil, []TxInput{txi}, []TxOutput{txo}}
	ts.hash()
	tss := []Transaction{ts}
	// 开始生成第一个区块
	bc.newGenesisBlockchain(tss)
	// 创世区块后，更新本地最新区块为1，向全网节点发送当前区块高度为1
	NewestBlockHeight = 1
	send.SendVersionToPeers(1)
	fmt.Println("已生成创世区块")
	// 重置utxo数据库，将创世数据存入

}

// 创建区块链
func (bc *blockchain) newGenesisBlockchain(transaction []Transaction) {
	// 判断一下是否已生成创世区块
	if len(bc.DB.View([]byte(LastBlockHashMapping), database.BlockBucket)) != 0 {
		log.Error("已生成创世区块")
	}

}

// 查找数据库中全部未花费的UTXO
func (bc *blockchain) findAllUTXOs() map[string][]*UTXO {
	utxosMap := make(map[string][]*UTXO)
	txInputmap := make(map[string][]TxInput)
	bcIterator := NewBlockchainIterator(bc)

	for {
		currentBlock := bcIterator.Next()
		if currentBlock == nil {
			return nil
		}
		// 必须倒序 否则有的已花费不会被扣掉
		/*代码首先遍历当前区块中的所有交易，这个遍历是以倒序的方式进行的，即从最后一个交易开始，
		一直到第一个交易。这是因为在区块链中，后发生的交易可能会花费前面交易的输出，
		所以倒序处理可以确保在处理每个交易输入时，之前已经处理过的交易不会被错误地认为是未花费的。 */
		for i := len(currentBlock.Transactions); i >= 0; i-- {
			var utxos = []*UTXO{}
			ts := currentBlock.Transactions[i]
			for _, input := range ts.Vint {
				txInputmap[string(input.TxHash)] = append(txInputmap[string(input.TxHash)], input)
			}
		VoutTag:
			for index, output := range ts.Vout {
				if txInputmap[string(ts.TxHash)] == nil {
					utxos = append(utxos, &UTXO{ts.TxHash, index, output})
				} else {
					for _, input := range txInputmap[string(ts.TxHash)] {
						if input.Index == index {
							continue VoutTag
						}
					}
					utxos = append(utxos, &UTXO{ts.TxHash, index, output})
				}
				utxosMap[string(ts.TxHash)] = utxos
			}
		}
		// 直到查到创世区块结束
		if isGenesisBlock(currentBlock) {
			break
		}
	}
	return utxosMap
}
