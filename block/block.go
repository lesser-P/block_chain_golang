package block

import "time"

type Block struct {
	// 上一个区块的hash
	PreHash []byte
	// 数据data
	Transactions []Transaction
	//时间戳
	TimeStamp int64
	//区块高度
	Height int
	//随机数
	Nonce int64
	//本区块hash
	Hash []byte
}

func mineBlock(transaction []Transaction, preHash []byte, height int) (*Block, error) {
	timestamp := time.Now().Unix()
	// hash数据+时间戳+上一个区块hash
	block := Block{
		PreHash:      preHash,
		Transactions: transaction,
		TimeStamp:    timestamp,
		Height:       height,
		Nonce:        0,
		Hash:         nil,
	}
}

func newGenesisBlock(transaction []Transaction) *blockchain {
	// 创世区块的上一个块hash默认设置成0数据
	preHash := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
}
