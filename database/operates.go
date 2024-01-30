package database

import (
	"github.com/boltdb/bolt"
	log "github.com/corgi-kx/logcustom"
)

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
