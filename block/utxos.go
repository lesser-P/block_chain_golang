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

// 同步数据,传入交易信息，保存输出信息到数据库，剔除输入信息
func (u *UTXOHandle) Synchrodata(tss []Transaction) {
	// 先将全部输入插入数据库
	for _, ts := range tss {
		utxos := []*UTXO{}
		for index, Vout := range ts.Vout {
			utxos = append(utxos, &UTXO{ts.TxHash, index, Vout})
		}
		u.BC.DB.Put(ts.TxHash, u.serialize(utxos), database.UTXOBucket)
	}
	// 用输出进行剔除
	for _, ts := range tss {
		for index, Vint := range ts.Vint {
			publicKeyHash := generatePublicKeyHash(Vint.PublicKey)
			// 遍历整个utxo数据库
			utxoByte := u.BC.DB.View(Vint.TxHash, database.UTXOBucket)
			if len(utxoByte) == 0 {
				log.Panic("synchrodata err : do not find utxo")
			}
			utxos := u.dserialize(utxoByte)
			newUTXO := []*UTXO{}
			for _, utxo := range utxos {
				// 如果条件都符合则说明这个utxo已经是被vinput消费了，所以跳过不计入新的utxo
				if utxo.Index == index && bytes.Equal(utxo.Vout.PublicKeyHash, publicKeyHash) {
					continue
				}
				newUTXO = append(newUTXO, utxo)
			}
			// 删除原本的utxo，保存剔除掉消费后的utxo
			u.BC.DB.Delete(Vint.TxHash, database.UTXOBucket)
			u.BC.DB.Put(Vint.TxHash, u.serialize(newUTXO), database.UTXOBucket)
		}
	}
}

// 获取数据库中为消费的utxo
func (u *UTXOHandle) findUTXOFromAddress(address string) []*UTXO {
	publicKeyHash := getPublicKeyHashFromAddress(address)
	utxosSlice := []*UTXO{}
	// 获取boly迭代器
	DBFileName := "blockchain" + ListenPort + ".db"
	db, err := bolt.Open(DBFileName, 0600, nil)
	if err != nil {
		log.Panic(err)
	}
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(database.UTXOBucket))
		if b == nil {
			return errors.New("datebase view err: not find bucket")
		}
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			utxos := u.dserialize(v)
			for _, utxo := range utxos {
				if bytes.Equal(utxo.Vout.PublicKeyHash, publicKeyHash) {
					utxosSlice = append(utxosSlice, utxo)
				}
			}
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	err = db.Close()
	if err != nil {
		log.Error("db close err:", err)
	}
	return utxosSlice
}
