package block

import (
	"block_chain_golang/database"
	"bytes"
	"encoding/gob"
	"errors"
	"github.com/boltdb/bolt"
	log "github.com/corgi-kx/logcustom"
)

/*
	utxo数据库创建的意义在于,不会每次进行转帐时遍历整个区块链,
	而是去utxo数据库查找未消费的交易输出,这样会大大降低性能问题
*/

type UTXOHandle struct {
	BC *blockchain
}

// 重置UTXO数据库
func (u *UTXOHandle) ResetUTXODataBase() {
	// 先查找全部未花费UTXO
	utxosMap := u.BC.findAllUTXOs()
	if utxosMap == nil {
		log.Info("找不到区块,暂不重置UTXO数据库")
		return
	}
	//删除旧的UTXO的数据库
	if database.IsBucketExist(u.BC.DB, database.UTXOBucket) {
		u.BC.DB.DeleteBucket(database.UTXOBucket)
	}
	//创建并将未花费的UTXO循环添加
	for k, v := range utxosMap {
		u.BC.DB.Put([]byte(k), u.serialize(v), database.UTXOBucket)
	}
}
func (u *UTXOHandle) serialize(utxos []*UTXO) []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(utxos)
	if err != nil {
		panic(err)
	}
	return result.Bytes()
}

func (u *UTXOHandle) dserialize(d []byte) []*UTXO {
	var model []*UTXO
	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&model)
	if err != nil {
		log.Panic(err)
	}
	return model
}

// 获取数据库中为消费的utxo
func (u *UTXOHandle) findUTXOFromAddress(address string) []*UTXO {
	publicKeyHash := getPublicKeyHashFromAddress(address)
	utxosSlic := []UTXO{}
	// 获取boly迭代器
	DBFileName := "blockchain" + ListenPort + ".db"
	db, err := bolt.Open(DBFileName, 0600, nil)
	if err != nil {
		log.Panic(err)
	}
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(database.UTXOBucket))
		if b == nil {
			return errors.New("datebase view err: not find bucket")
		}
		cursor := b.Cursor()
	})
}
