package block

import (
	"bytes"
	"encoding/gob"
	log "github.com/corgi-kx/logcustom"
	"math/big"
	"time"
)

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

// 进行挖矿来生成区块
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
	pow := NewProofOfWork(&block)
	nonce, hash, err := pow.run()
	if err != nil {
		return nil, err
	}
	block.Nonce = nonce
	block.Hash = hash[:]
	log.Info("pow Verify:", pow.Verify())
	log.Infof("以生成新的区块，区块高度为%d", block.Height)
	return &block, nil
}

// 生成创世区块
func newGenesisBlock(transaction []Transaction) *Block {
	// 创世区块的上一个块hash默认设置成0数据
	preHash := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	genesisBlock, err := mineBlock(transaction, preHash, 1)
	if err != nil {
		log.Error(err)
	}
	return genesisBlock
}

// 判断是否是创世区块
func isGenesisBlock(block *Block) bool {
	var hashInt big.Int
	hashInt.SetBytes(block.PreHash)
	if hashInt.Cmp(big.NewInt(0)) == 0 {
		return true
	}
	return false
}

func (b *Block) Serialize() []byte {
	var result bytes.Buffer
	err := gob.NewEncoder(&result).Encode(b)
	if err != nil {
		panic(err)
	}
	return result.Bytes()
}

func (b *Block) Deserialize(d []byte) {
	err := gob.NewDecoder(bytes.NewReader(d)).Decode(b)
	if err != nil {
		panic(err)
	}
}
