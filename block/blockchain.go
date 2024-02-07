package block

import (
	"block_chain_golang/database"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/corgi-kx/logcustom"
	"math/big"
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
		u := UTXOHandle{bc}
		utxos := u.findUTXOFromAddress(fromAddress)
		if len(utxos) == 0 {
			log.Error("余额为零，无法进行转账交易")
			return
		}

		// 将utxos添加尚未打包进区块的交易信息
		if tss != nil {
			for _, ts := range tss {
			tageVout:
				for index, vOut := range ts.Vout {
					// 交易输出者的公钥哈希和发起交易者公钥哈希比较
					if bytes.Compare(vOut.PublicKeyHash, generatePublicKeyHash(fromKeys.PublicKey)) != 0 {
						// 如果不是同一个地址那么跳过
						continue
					}
					for _, utxo := range utxos {
						if bytes.Equal(ts.TxHash, utxo.Hash) && utxo.Index == index {
							// 如果已经添加过了就不再添加
							continue tageVout
						}
					}
					utxos = append(utxos, &UTXO{ts.TxHash, index, vOut})
				}
				// 剔除已经花费的utxo
				for _, vInt := range ts.Vint {
					for index, utxo := range utxos {
						if bytes.Equal(vInt.TxHash, utxo.Hash) && vInt.Index == index {
							utxos = append(utxos[:index], utxos[index+1:]...)
						}
					}
				}
			}
		}

		// 打包交易的核心操作
		newTXInput := []TxInput{}
		newTXOutput := []TxOutput{}
		var amount int
		for _, utxo := range utxos {
			amount = amount + utxo.Vout.Value
			// 输入交易的交易哈希是未消费输出交易的哈希
			newTXInput = append(newTXInput, TxInput{utxo.Hash, utxo.Index, nil, fromKeys.PublicKey})
			if amount > amountSlice[index] {
				tfrom := TxOutput{}
				tfrom.Value = amount - amountSlice[index]
				tfrom.PublicKeyHash = generatePublicKeyHash(fromKeys.PublicKey)
				tTo := TxOutput{}
				tTo.Value = amountSlice[index]
				tTo.PublicKeyHash = toKeysPublicKeyHash
				newTXOutput = append(newTXOutput, tfrom)
				newTXOutput = append(newTXOutput, tTo)
				break
			} else if amount == amountSlice[index] {
				tTo := TxOutput{}
				tTo.Value = amountSlice[index]
				tTo.PublicKeyHash = toKeysPublicKeyHash
				newTXOutput = append(newTXOutput, tTo)
				break
			}
		}
		//如果余额不足则会跳过不会打包进入交易
		if amount < amountSlice[index] {
			log.Error("余额不足，无法进行转账交易")
			continue
		}
		ts := Transaction{nil, newTXInput, newTXOutput[:]}
		ts.hash()
		tss = append(tss, ts)
	}
	if tss == nil {
		return
	}
	// 对交易进行签名
	bc.signatureTransactions(tss, wallets)
	send.SendTransToPeers(tss)
}

// 校验交易余额是否足够，如果不够则剔除
func (bc *blockchain) VerifyTransBalance(tss *[]Transaction) {
	// 获取每个地址的UTXO余额，并存入字典
	var balance = map[string]int{}
	for i := range *tss {
		fromAddress := GetAddressFromPublicKey((*tss)[i].Vint[0].PublicKey)
		// 获取数据库中的utxo
		u := UTXOHandle{bc}
		utxos := u.findUTXOFromAddress(fromAddress)
		if len(utxos) == 0 {
			log.Warnf("%s 余额为0！", fromAddress)
			continue
		}
		amount := 0
		for _, v := range utxos {
			amount += v.Vout.Value
		}
		balance[fromAddress] = amount
	}

circle:
	for i := range *tss {
		fromAddress := GetAddressFromPublicKey((*tss)[i].Vint[0].PublicKey)
		u := UTXOHandle{bc}
		// 查看这个地址的所有未交易输出
		utxos := u.findUTXOFromAddress(fromAddress)
		var utxoAmount int //将要花费的utxo
		var voutAmount int //vout剩余的utxo
		var costAmount int //vint将要花费的总utxo减去vout剩余的utxo等于花费的钱数
		//获取每笔vin的值
		for _, vIn := range (*tss)[i].Vint {
			for _, vUTXO := range utxos {
				if bytes.Equal(vIn.TxHash, vUTXO.Hash) && vIn.Index == vUTXO.Index {
					utxoAmount += vUTXO.Vout.Value
				}
			}
		}
		for _, vOut := range (*tss)[i].Vout {
			if bytes.Equal(getPublicKeyHashFromAddress(fromAddress), vOut.PublicKeyHash) {
				voutAmount += vOut.Value
			}
		}
		costAmount = utxoAmount - voutAmount
		if _, ok := balance[fromAddress]; ok {
			balance[fromAddress] -= costAmount
			if balance[fromAddress] < 0 {
				log.Errorf("%s 余额不够，已将此笔交易剔除")
				*tss = append((*tss)[:i], (*tss)[i+1:]...)
				balance[fromAddress] += costAmount
				goto circle
			}
		} else {
			log.Errorf("%s 余额不够，已将此笔交易剔除")
			*tss = append((*tss)[:i], (*tss)[i+1:]...)
			goto circle
		}
	}
	log.Debug("已完成UTXO交易余额验证")
}

// 交易转账
func (bc *blockchain) Transfer(tss []Transaction, send Sender) {
	// 如果是创世区块的交易则无需进行数字签名验证
	if !isGenesisTransaction(tss) {
		// 交易的数字签名验证
		bc.verifyTransactionsSign(&tss)
		if len(tss) == 0 {
			log.Errorf("没有通过的数字签名验证，不予挖矿出块")
			return
		}
		//进行余额验证

		//如果设置了奖励地址，则挖矿成功后给予奖励代币

	}
}

// 根据交易id查找对应的交易信息
func (bc *blockchain) findTransaction(tss []Transaction, ID []byte) (Transaction, error) {
	//先查找未插入数据库的交易
	if len(tss) != 0 {
		for _, tx := range tss {
			if bytes.Compare(tx.TxHash, ID) == 0 {
				return tx, nil
			}
		}
	}
	bci := NewBlockchainIterator(bc)
	// 在查找数据库中存在的交易
	for {
		// 获取当前区块
		block := bci.Next()
		for _, tx := range block.Transactions {
			if bytes.Compare(tx.TxHash, ID) == 0 {
				return tx, nil
			}
		}
		// 一只迭代到创世区块然后结束
		var hashInt big.Int
		hashInt.SetBytes(block.PreHash)
		if big.NewInt(0).Cmp(&hashInt) == 0 {
			break
		}
	}
	return Transaction{}, errors.New("FindTransaction err : Transaction is not found")
}

// 签名
func (bc *blockchain) signatureTransactions(tss []Transaction, wallets *wallets) {
	for i := range tss {
		// 获得这笔交易的拷贝
		copyTs := tss[i].customCopy()

		for index := range tss[i].Vint {
			bk := bitcoinKeys{nil, tss[i].Vint[index].PublicKey, nil}
			address := bk.getAddress()
			//从数据库或者为打包进数据库的交易数组中，找到vint所对应的交易信息
			trans, err := bc.findTransaction(tss, tss[i].Vint[index].TxHash)
			if err != nil {
				log.Fatal(err)
			}
			copyTs.Vint[index].Signature = nil
			// 将拷贝后的交易里面的公钥替换为公钥hash
			copyTs.Vint[index].PublicKey = trans.Vout[tss[i].Vint[index].Index].PublicKeyHash
			// 对拷贝后的交易进行整体hash
			copyTs.TxHash = copyTs.hashSign()
			copyTs.Vint[index].PublicKey = nil
			privKey := wallets.Wallets[string(address)].PrivateKey
			// 进行签名操作
			tss[i].Vint[index].Signature = ellipticCurveSign(privKey, copyTs.TxHash)
		}
	}
}

// 数字签名验证
func (bc *blockchain) verifyTransactionsSign(tss *[]Transaction) {
circle:
	for i := range *tss {
		copyTs := (*tss)[i].customCopy()
		for index, Vin := range (*tss)[i].Vint {
			findTs, err := bc.findTransaction(*tss, Vin.TxHash)
			if err != nil {
				log.Fatal(err)
			}
			// 先验证输入地址的公钥hash与指定的utxo输出的公钥hash是否相同
			// 因为只有正确的私钥才能生成与UTXO公钥哈希匹配的公钥哈希
			// 输入交易的TXHash就是输出交易的TXHash，为了溯源，而输入交易的TXHash则是由持有者的私钥生成
			if !bytes.Equal(findTs.Vout[Vin.Index].PublicKeyHash, generatePublicKeyHash(Vin.PublicKey)) {
				log.Errorf("签名验证失败 %x 笔交易的vin并非是本人", (*tss)[i].TxHash)
				// 验证失败的话删除错误的交易再重新验证所有交易
				*tss = append((*tss)[:i], (*tss)[i+1:]...)
				goto circle
			}
			copyTs.Vint[index].Signature = nil
			copyTs.Vint[index].PublicKey = findTs.Vout[Vin.Index].PublicKeyHash
			copyTs.TxHash = copyTs.hashSign()
			copyTs.Vint[index].PublicKey = nil
			// 进行签名
			if !ellipticCurveVerify(Vin.PublicKey, Vin.Signature, copyTs.TxHash) {
				log.Errorf("此笔交易：%x没通过签名验证", (*tss)[i].TxHash)
				// 跳过这笔交易
				*tss = append((*tss)[:i], (*tss)[i+1:]...)
				goto circle
			}
		}
	}
	log.Debug("已完成数字签名验证")
}
