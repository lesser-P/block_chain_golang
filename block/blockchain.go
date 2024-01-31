package block

import (
	"block_chain_golang/database"
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
	bc.
}

// 创建区块链
func (bc *blockchain)newGenesisBlockchain(transaction []Transaction)  {
	// 判断一下是否已生成创世区块
	if len(bc.DB.View([]byte(LastBlockHashMapping),database.BlockBucket))!=0 {
		log.Error("已生成创世区块")
	}

}