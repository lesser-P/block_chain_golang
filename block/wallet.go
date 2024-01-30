package block

import (
	"bytes"
	"encoding/gob"
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
