package block

import (
	"block_chain_golang/database"
	"bytes"
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

// 生成钱包
func (w *wallets) GenerateWallet(bd *database.BlockchainDB, keys func(s []string) *bitcoinKeys, s []string) (address, privKey, mnemonicWord string) {
	// 根据传入的函数来决定创建bitcoinKeys的策略
	bitcoinKeys := keys(s)
	if bitcoinKeys == nil {
		log.Fatal("创建钱包失败，检查助记词是否符合创建原则")
	}
	privKey = bitcoinKeys.GetPrivateKey()
	addressByte := bitcoinKeys.getAddress()
	w.storage(addressByte, bitcoinKeys, bd)
	// 将地址存入实例
	address = string(addressByte)
	// 将助记词拼接成json格式并返回
	mnemonicWord = "["
	for i, v := range bitcoinKeys.MnemonicWord {
		mnemonicWord += "\"" + v + "\""
		if i != len(bitcoinKeys.MnemonicWord)-1 {
			mnemonicWord += ","
		} else {
			mnemonicWord += "]"
		}
	}
	return
}

// 将钱包信息存入数据库
func (w *wallets) storage(address []byte, keys *bitcoinKeys, bd *database.BlockchainDB) {
	b := bd.View(address, database.AddrBucket)
	if len(b) != 0 {
		log.Warn("钱包已存在")
		return
	}
	// 将公私钥以地址为键 存入数据库
	bd.Put(address, keys.serliazle(), database.AddrBucket)

	// 将地址存入地址导航
	listBytes := bd.View([]byte(addrListMapping), database.AddrBucket)
	if len(listBytes) == 0 {
		a := addressList{address}
		bd.Put([]byte(addrListMapping), a.serliazle(), database.AddrBucket)
	} else {
		addressList := addressList{}
		// 把listByte中的内容反序列化到addressList中
		addressList.Deserialize(listBytes)
		addressList = append(addressList, address)
		bd.Put([]byte(addrListMapping), addressList.serliazle(), database.AddrBucket)
	}
}
