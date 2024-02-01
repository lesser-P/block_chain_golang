package block

import "block_chain_golang/database"

type blockchainIterator struct {
	CurrentBlockHash []byte
	BD               *database.BlockchainDB
}

// 获取区块迭代器实例
func NewBlockchainIterator(bc *blockchain) *blockchainIterator {
	iterator := blockchainIterator{
		CurrentBlockHash: bc.DB.View([]byte(LastBlockHashMapping), database.BlockBucket),
		BD:               bc.DB,
	}
	return &iterator
}

// 迭代下一个区块信息
func (bi *blockchainIterator) Next() *Block {
	currentByte := bi.BD.View(bi.CurrentBlockHash, database.BlockBucket)
	if len(currentByte) == 0 {
		return nil
	}
	block := Block{}
	// 反序列化获得当前区块信息
	block.Deserialize(currentByte)
	bi.CurrentBlockHash = block.PreHash
	return &block
}
