package network

import (
	"block_chain_golang/block"
	"bytes"
	"encoding/gob"
	log "github.com/corgi-kx/logcustom"
)

type Transaction struct {
	TxHash []byte
	//UTXO输入
	Vint []block.TxInput
	//UTXO输出
	Vout []block.TxOutput

	AddrFrom string //交易发起地址
}
type Transactions struct {
	Ts []Transaction
}

// 序列化
func (t *Transactions) Serialize() []byte {
	var result bytes.Buffer
	err := gob.NewEncoder(&result).Encode(t) //序列化为字节切片
	if err != nil {
		panic(err)
	}
	return result.Bytes()
}

// 反序列化
func (t *Transactions) Deserialize(d []byte) {
	err := gob.NewDecoder(bytes.NewReader(d)).Decode(t)
	if err != nil {
		log.Panic(err)
	}
}
