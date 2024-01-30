package block

import (
	"block_chain_golang/database"
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	log "github.com/corgi-kx/logcustom"
)

type addressList [][]byte

// 序列化地址列表
func (a *addressList) serliazle() []byte {
	var result bytes.Buffer
	err := gob.NewEncoder(&result).Encode(a)
	if err != nil {
		panic(err)
	}
	return result.Bytes()
}

// 反序列化地址列表
func (v *addressList) Deserialize(d []byte) {
	err := gob.NewDecoder(bytes.NewReader(d)).Decode(v)
	if err != nil {
		panic(err)
	}
}

type wallets struct {
	Wallets map[string]*bitcoinKeys
}

// 创建一个新钱包实例
func NewWallets(bd *database.BlockchainDB) *wallets {
	w := &wallets{make(map[string]*bitcoinKeys)}
	//如果钱包存在，则先取出地址信息，根据地址信息取出钱包信息
	if database.IsBucketExist(bd, database.AddrBucket) {
		addressList := GetAllAddress(bd)
		if addressList == nil {
			return w
		}
		for _, v := range *addressList {
			key := bitcoinKeys{}
			key.Deserialize(bd.View(v, database.AddrBucket))
			w.Wallets[string(v)] = &key
		}
		return w
	}
	return w
}

// 序列化
func (b *bitcoinKeys) serliazle() []byte {
	var result bytes.Buffer
	gob.Register(elliptic.P256())
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(b)
	if err != nil {
		panic(err)
	}
	return result.Bytes()
}

// 反序列化
func (v *bitcoinKeys) Deserialize(d []byte) {
	decoder := gob.NewDecoder(bytes.NewReader(d))
	gob.Register(elliptic.P256())
	err := decoder.Decode(v)
	if err != nil {
		log.Error(err)
	}
}

// 获得所有地址
func GetAllAddress(bd *database.BlockchainDB) *addressList {
	listaddress := bd.View([]byte(addrListMapping), database.AddrBucket)
	if len(listaddress) == 0 {
		return nil
	}
	add := &addressList{}
	// 反序列化到addressList
	add.Deserialize(listaddress)
	return add
}
