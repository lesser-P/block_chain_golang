package database

import (
	"github.com/boltdb/bolt"
	log "github.com/corgi-kx/logcustom"
)

// 存入数据
func (bd *BlockchainDB) Put(k, v []byte, bt BucketType) {
	var blockDBFileName = "blockchain_" + ListenPort + ".db"
	db, err := bolt.Open(blockDBFileName, 0600, nil)
	defer db.Close()
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bt))
		if bucket == nil {
			bucket, err = tx.CreateBucket([]byte(bt))
			if err != nil {
				log.Panic(err)
			}
		}
		err := bucket.Put(k, v)
		if err != nil {
			log.Panic(err)
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

func (bd *BlockchainDB) View(k []byte, bt BucketType) []byte {
	var blockDBFileName = "blockchain_" + ListenPort + ".db"
	db, err := bolt.Open(blockDBFileName, 0600, nil)
	defer db.Close()
	if err != nil {
		return nil
	}
	var result []byte
	err = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bt))
		if bucket == nil {
			log.Error("没有对应的仓库")
		}
		result = bucket.Get(k)
		return nil
	})
	if err != nil {
		return nil
	}
	return result
}

// 删除仓库
func (bd *BlockchainDB) DeleteBucket(bt BucketType) bool {
	var DBFileName = "blockchain_" + ListenPort + ".db"
	db, err := bolt.Open(DBFileName, 0600, nil)
	defer db.Close()
	if err != nil {
		log.Panic(err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		err := tx.DeleteBucket([]byte(bt))
		if err != nil {
			log.Panic(err)
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return true
}
