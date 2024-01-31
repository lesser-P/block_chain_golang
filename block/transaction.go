package block

import (
	"block_chain_golang/util"
	"bytes"
	"encoding/gob"
	"github.com/cloudflare/cfssl/scan/crypto/sha256"
	log "github.com/corgi-kx/logcustom"
)

// 交易列表信息
type Transaction struct {
	TxHash []byte
	//UTXO输入
	Vint []TxInput
	//UTXO输出
	Vout []TxOutput
}

// 对此笔交易的输入输出进行hash运算后存入交易hash(txhash)
func (t *Transaction) hash() {
	tBytes := t.Serialize()
	//加入随机数byte
	randomNumber := util.GenerateRealRandom()
	randomByte := util.Int64ToBytes(randomNumber)
	sumByte := bytes.Join([][]byte{tBytes, randomByte}, []byte(""))
	//再次哈希
	hashByte := sha256.Sum256(sumByte)
	t.TxHash = hashByte[:]
}

// 数字签名的hash
func (t *Transaction) hashSign() []byte {
	t.TxHash = nil
	nHash := []byte{}
	for _, input := range t.Vint {
		nHash = append(nHash, input.TxHash...)
		nHash = append(nHash, input.PublicKey...)
		nHash = append(nHash, util.Int64ToBytes(int64(input.Index))...)
	}
	for _, output := range t.Vout {
		nHash = append(nHash, output.PublicKeyHash...)
		nHash = append(nHash, util.Int64ToBytes(int64(output.Value))...)
	}
	hashByte := sha256.Sum256(nHash)
	return hashByte[:]
}

// 将transaction序列化成[]byte
func (t *Transaction) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(t)
	if err != nil {
		panic(err)
	}
	return result.Bytes()
}

// 把整笔交易里的成员依次转换成字节数组，拼接成整体后返回
func (t *Transaction) getTransBytes() []byte {
	if t.TxHash == nil || t.Vout == nil {
		log.Panic("交易信息不完整，无法拼接成字节数组")
		return nil
	}
	transBytes := []byte{}
	transBytes = append(transBytes, t.TxHash...)
	for _, input := range t.Vint {
		transBytes = append(transBytes, input.TxHash...)
		transBytes = append(transBytes, input.PublicKey...)
		transBytes = append(transBytes, input.Signature...)
		transBytes = append(transBytes, util.Int64ToBytes(int64(input.Index))...)
	}
	for _, output := range t.Vout {
		transBytes = append(transBytes, util.Int64ToBytes(int64(output.Value))...)
		transBytes = append(transBytes, output.PublicKeyHash...)
	}
	return transBytes
}

// 从原交易里拷贝出一个新的交易
func (t *Transaction) customCopy() Transaction {
	newVint := []TxInput{}
	newVout := []TxOutput{}

	for _, input := range t.Vint {
		newVint = append(newVint, TxInput{input.TxHash, input.Index, nil, nil})
	}

	for _, output := range t.Vout {
		newVout = append(newVout, TxOutput{output.Value, output.PublicKeyHash})
	}
	return Transaction{t.TxHash, newVint, newVout}
}

// 判断是否是创世区块的交易
func isGenesisTransaction(tss []Transaction) bool {
	if tss != nil {
		if tss[0].Vint[0].Index == -1 {
			return true
		}
	}
	return false
}
