package block

import (
	"block_chain_golang/util"
	"bytes"
	"crypto/rand"
	"errors"
	"github.com/cloudflare/cfssl/scan/crypto/sha256"
	log "github.com/corgi-kx/logcustom"
	"math"
	"math/big"
	"time"
)

// 工作量证明算法

// pow结构体
type proofOfWork struct {
	*Block
	Target *big.Int
}

// 获取POW实例
func NewProofOfWork(block *Block) *proofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, 256-TargetBits)
	pow := &proofOfWork{block, target}
	return pow
}

// 进行hash运算，获取当前区块的hash值
func (p *proofOfWork) run() (int64, []byte, error) {
	var nonce int64 = 0
	var hashByte [32]byte
	var hashInt big.Int
	log.Info("准备挖矿...")
	// 开启一个计数器，每隔五秒打印一下当前挖矿，用来直观展现挖矿情况
	times := 0
	ticker := time.NewTicker(5 * time.Second)

	go func(t *time.Ticker) {
		for {
			// 阻塞操作会等待timeticker发送一个信号
			<-t.C
			times += 5
			log.Infof("正在挖矿，挖矿区块高度为%d，已经运行%ds，nonce值：%d，当前hash：%x", p.Height, times, nonce, hashByte)
		}
	}(ticker)

	for nonce < maxInt {
		//监测网络上其他节点是否已经挖出了区块
		if p.Height <= NewestBlockHeight {
			// 结束计数器
			ticker.Stop()
			return 0, nil, errors.New("检测到当前节点已接收到最新区块，所以终止此块的挖矿操作")
		}
		data := p.jointData(nonce)
		// 上一区块所有数据的hash
		hashByte = sha256.Sum256(data)
		hashInt.SetBytes(hashByte[:])
		/*如果hash后的data值小于设置的挖矿难度大数字，则代表挖矿成功
		在这段代码中，hash后的data值小于设置的挖矿难度大数字是指，通过对区块数据进行哈希运算，
		得到的哈希值（这里被视为一个大整数）如果小于预设的目标值（也是一个大整数，代表挖矿难度），
		那么就认为挖矿成功。  这个挖矿过程实际上是在寻找一个满足特定条件的哈希值，
		这个条件就是哈希值必须小于预设的目标值。这个过程通常需要大量的计算，因此被称为“挖矿”。
		当找到这样一个哈希值时，就认为挖矿成功，新的区块就可以添加到区块链中*/
		if hashInt.Cmp(p.Target) == -1 {
			break
		} else {
			// nonce++
			bigInt, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
			if err != nil {
				log.Error("随机数错误：", err)
			}
			nonce = bigInt.Int64()
		}
	}
	//结束计时器
	ticker.Stop()
	log.Infof("本节点已成功挖到区块！！，高度：%d，nonce的值为：%d，区块hash为%x", p.Height, nonce, hashByte)
	return nonce, hashByte[:], nil
}

// 将上一区块的hash，数据，时间戳，难度位数，随机数，拼接成字节数组
func (p *proofOfWork) jointData(nonce int64) (data []byte) {
	preHash := p.Block.PreHash
	timeStampByte := util.Int64ToBytes(time.Now().Unix())
	heightByte := util.Int64ToBytes(int64(p.Block.Height))
	nonceByte := util.Int64ToBytes(nonce)
	targetBitsByte := util.Int64ToBytes(int64(TargetBits))
	// 拼接成交易数组
	transData := [][]byte{}
	for _, v := range p.Block.Transactions {
		transData = append(transData, v.getTransBytes()) // 不使用gob是因为gob同样的数据序列化后的字节数组可能数据不一致
	}
	// 获取交易数据的根默克尔节点
	mr := util.NewMerkelTree(transData)

	data = bytes.Join([][]byte{
		preHash, timeStampByte, heightByte, mr.MerkelRootNode.Data, nonceByte, targetBitsByte,
	}, []byte(""))
	return data
}

func (p *proofOfWork) Verify() bool {
	/*TargetBits是难度指标，它表示目标哈希值的前TargetBits位必须是0，
	剩下的256-TargetBits位可以是任意值*/
	target := big.NewInt(1)
	target.Lsh(target, 256-TargetBits)
	data := p.jointData(p.Block.Nonce)
	hash := sha256.Sum256(data)
	var hashInt big.Int
	hashInt.SetBytes(hash[:])
	if hashInt.Cmp(target) == -1 {
		return true
	}
	return false
}
