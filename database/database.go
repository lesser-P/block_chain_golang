package database

import (
	"github.com/boltdb/bolt"
	log "github.com/corgi-kx/logcustom"
)

// 监听地址
var ListenPort string

// 仓库类型
type BucketType string

const (
	BlockBucket BucketType = "blocks"
	AddrBucket  BucketType = "address"
	UTXOBucket  BucketType = "utxo"
)

type BlockchainDB struct {
	ListenPort string
}

func New() *BlockchainDB {
	block := &BlockchainDB{ListenPort}
	return block
}

// 判断仓库是否存在
func IsBucketExist(bd *BlockchainDB, bt BucketType) bool {
	var isBucketExist bool
	DBFileName := "blockchain_" + ListenPort + ".db"
	db, err := bolt.Open(DBFileName, 0600, nil)
	if err != nil {
		log.Error(err)
	}
	// 事务，只读操作
	err = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bt))
		if bucket == nil {
			isBucketExist = false
		} else {
			isBucketExist = true
		}
		return nil
	})
	if err != nil {
		log.Error(err)
	}
	db.Close()
	return isBucketExist
}
