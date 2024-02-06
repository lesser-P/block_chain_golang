package block

import (
	"block_chain_golang/database"
	"encoding/json"
	"fmt"
	log "github.com/corgi-kx/logcustom"
)

type blockchain struct {
	DB *database.BlockchainDB //封装blot结构体
}

func NewBlockchain() *blockchain {
	return &blockchain{database.New()}
}

// 创建创世区块交易信息
func (bc *blockchain) CreateGenesisTransaction(address string, value int, send Sender) {
	// 判断地址格式是否正确
	if !IsVailBitcoinAddress(address) {
		log.Error("地址错误", address)
		return
	}

	//创世区块数据
	txi := TxInput{
		[]byte{},
		-1,
		nil,
		nil,
	}

	// 本地存创世区块的公私钥信息
	wallets := NewWallets(bc.DB)
	//创世区块地址的公私钥信息
	genesisKeys, ok := wallets.Wallets[address]
	if !ok {
		log.Error("没有找到对应地址的公私钥信息")
	}
	publicKeyHash := generatePublicKeyHash(genesisKeys.PublicKey)
	txo := TxOutput{Value: value, PublicKeyHash: publicKeyHash}
	ts := Transaction{nil, []TxInput{txi}, []TxOutput{txo}}
	ts.hash()
	tss := []Transaction{ts}
	// 开始生成第一个区块
	bc.newGenesisBlockchain(tss)
	// 创世区块后，更新本地最新区块为1，向全网节点发送当前区块高度为1
	NewestBlockHeight = 1
	send.SendVersionToPeers(1)
	fmt.Println("已生成创世区块")
	// 重置utxo数据库，将创世数据存入
	utxos := UTXOHandle{bc}
	utxos.ResetUTXODataBase()
}

// 创建区块链
func (bc *blockchain) newGenesisBlockchain(transaction []Transaction) {
	// 判断一下是否已生成创世区块
	if len(bc.DB.View([]byte(LastBlockHashMapping), database.BlockBucket)) != 0 {
		log.Error("已生成创世区块")
	}
	// 生成创世区块
	genesisBlock := newGenesisBlock(transaction)
	// 添加到区块数据库
	bc.AddBlock(genesisBlock)

}

// 添加区块信息到数据库，并更新lastHash
func (bc *blockchain) AddBlock(block *Block) {
	bc.DB.Put(block.Hash, block.Serialize(), database.BlockBucket)
	iterator := NewBlockchainIterator(bc)
	// 获得当前区块
	currentBlock := iterator.Next()
	if currentBlock == nil || currentBlock.Height < block.Height {
		// 这个更像是一个更新最新区块的操作
		bc.DB.Put([]byte(LastBlockHashMapping), block.Hash, database.BlockBucket)
	}

}

// 查找数据库中全部未花费的UTXO
func (bc *blockchain) findAllUTXOs() map[string][]*UTXO {
	utxosMap := make(map[string][]*UTXO)
	txInputmap := make(map[string][]TxInput)
	bcIterator := NewBlockchainIterator(bc)

	for {
		currentBlock := bcIterator.Next()
		if currentBlock == nil {
			return nil
		}
		// 必须倒序 否则有的已花费不会被扣掉
		/*代码首先遍历当前区块中的所有交易，这个遍历是以倒序的方式进行的，即从最后一个交易开始，
		一直到第一个交易。这是因为在区块链中，后发生的交易可能会花费前面交易的输出，
		所以倒序处理可以确保在处理每个交易输入时，之前已经处理过的交易不会被错误地认为是未花费的。 */
		for i := len(currentBlock.Transactions); i >= 0; i-- {
			var utxos = []*UTXO{}
			ts := currentBlock.Transactions[i]
			for _, input := range ts.Vint {
				txInputmap[string(input.TxHash)] = append(txInputmap[string(input.TxHash)], input)
			}
		VoutTag:
			for index, output := range ts.Vout {
				// 当前交易的hash没有输入说明这个输出没有被任何输入引用，因此他是一个未花费的输出，将其添加到utxos切片中。
				if txInputmap[string(ts.TxHash)] == nil {
					utxos = append(utxos, &UTXO{ts.TxHash, index, output})
				} else {
					// 有输入数据则遍历这些输入
					for _, input := range txInputmap[string(ts.TxHash)] {
						//如果相同则说明这个输出已经被花费了，不需要添加到utxos切片中
						if input.Index == index {
							continue VoutTag
						}
					}
					utxos = append(utxos, &UTXO{ts.TxHash, index, output})
				}
				utxosMap[string(ts.TxHash)] = utxos
			}
		}
		// 直到查到创世区块结束
		if isGenesisBlock(currentBlock) {
			break
		}
	}
	return utxosMap
}

// 创建挖矿奖励地址交易
func (bc *blockchain) CreateRewardTransaction(address string) Transaction {
	if address == "" {
		log.Warnf("没有设置挖矿奖励地址，挖矿成功则不会产生奖励")
		return Transaction{}
	}
	// 判断地址格式是否正确
	if !IsVailBitcoinAddress(address) {
		log.Warnf("地址格式不正确%s", address)
		return Transaction{}
	}
	publicKeyHashByte := getPublicKeyHashFromAddress(address)
	txo := TxOutput{
		Value:         TokenRewardNum,
		PublicKeyHash: publicKeyHashByte,
	}
	tx := Transaction{nil, nil, []TxOutput{txo}}
	tx.hash()
	return tx
}

// 创建UXTO交易实例
func (bc *blockchain) CreateTransaction(from, to string, amount string, send Sender) {
	// 判断是否已生成创世区块
	if len(bc.DB.View([]byte(LastBlockHashMapping), database.BlockBucket)) == 0 {
		log.Error("还没有生成创世区块，不可进行转账操作！")
		return
	}
	// 判断是否设置了挖矿地址，没设置的话会给出提示
	if len(bc.DB.View([]byte(RewardAddrMapping), database.AddrBucket)) == 0 {
		log.Error("没有设置挖矿地址，如果挖出区块将不会给予奖励代币!")
	}

	fromSlice := []string{}
	toSlice := []string{}
	amountSlice := []int{}

	// 对传入的信息进行交验检测
	err := json.Unmarshal([]byte(from), &fromSlice)
	if err != nil {
		log.Error("json err", err)
		return
	}
	err = json.Unmarshal([]byte(to), &toSlice)
	if err != nil {
		log.Error("json err", err)
		return
	}
	err = json.Unmarshal([]byte(amount), &amountSlice)
	if err != nil {
		log.Error("json err", err)
		return
	}
	if len(fromSlice) != len(toSlice) || len(fromSlice) != len(amountSlice) {
		log.Error("转账数组长度不一致")
		return
	}

	for i, v := range fromSlice {
		if !IsVailBitcoinAddress(v) {
			log.Error("地址格式不正确已将此笔交易删除", v)
			if i < len(fromSlice)-1 {
				fromSlice = append(fromSlice[:i], fromSlice[i+1:]...)
				toSlice = append(toSlice[:i], toSlice[i+1:]...)
				amountSlice = append(amountSlice[:i], amountSlice[i+1:]...)
			} else {
				// 最后一个直接删除
				fromSlice = append(fromSlice[:i])
				toSlice = append(toSlice[:i])
				amountSlice = append(amountSlice[:i])
			}
		}
	}

	for i, v := range toSlice {
		if !IsVailBitcoinAddress(v) {
			log.Error("地址格式不正确已将此笔交易删除", v)
			if i < len(toSlice)-1 {
				fromSlice = append(fromSlice[:i], fromSlice[i+1:]...)
				toSlice = append(toSlice[:i], toSlice[i+1:]...)
				amountSlice = append(amountSlice[:i], amountSlice[i+1:]...)
			} else {
				fromSlice = append(fromSlice[:i])
				toSlice = append(toSlice[:i])
				amountSlice = append(amountSlice[:i])
			}
		}
	}

	for i, v := range amountSlice {
		if v < 0 {
			log.Error("转账金额不能为负数")
			if i < len(amountSlice)-1 {
				fromSlice = append(fromSlice[:i], fromSlice[i+1:]...)
				toSlice = append(toSlice[:i], toSlice[i+1:]...)
				amountSlice = append(amountSlice[:i], amountSlice[i+1:]...)
			} else {
				fromSlice = append(fromSlice[:i])
				toSlice = append(toSlice[:i])
				amountSlice = append(amountSlice[:i])
			}
		}
	}

	var tss []Transaction
	wallets := NewWallets(bc.DB)
	for index, fromAddress := range fromSlice {
		fromKeys, ok := wallets.Wallets[fromAddress]
		if !ok {
			log.Error("没有找到地址所对应的公钥，跳过此笔交易")
			continue
		}
		toKeysPublicKeyHash := getPublicKeyHashFromAddress(toSlice[index])

		if fromAddress == toSlice[index] {
			log.Error("转账地址不能相同")
			return
		}
		u := UTXOHandle{
			bc,
		}
		u.
	}

}
