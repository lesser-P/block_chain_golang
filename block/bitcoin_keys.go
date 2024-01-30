package block

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	log "github.com/corgi-kx/logcustom"
	"math/big"
	"os"
)

type bitcoinKeys struct {
	PrivateKey   *ecdsa.PrivateKey
	PublicKey    []byte
	MnemonicWord []string
}

// 创建公私钥实例
func NewBitcoinKeys(nothing []string) *bitcoinKeys {
	b := &bitcoinKeys{
		PrivateKey: nil,
		PublicKey:  nil,
	}
	b.MnemonicWord = getChineseMnemonicWord()
	b.newKeypair()
	return b
}

// 创建中文助记词
func getChineseMnemonicWord() []string {
	file, err := os.Open(ChineseMnwordPath)
	defer file.Close()
	if err != nil {
		log.Error(err)
	}
	s := []string{}
	//因为种子最高40位，所以就取7对词语 7*2*3=42，返回后再截取40位
	for i := 0; i < 7; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(5948)) //词库一共5948对词语，设置最高随机数为5948
		if err != nil {
			log.Error(err)
		}
		b := make([]byte, 6)
		//每一对词语在文件中占据7个字节的空间（包括一个分隔符），所以 n*7 确保了每次读取都是从某一对词语的开始位置读起。+3 是为了跳过每对词语前面的分隔符。
		_, err = file.ReadAt(b, n.Int64()*7+3)
		if err != nil {
			log.Error(err)
		}
		s = append(s, string(b))
	}
	return s
}

// 根据中文助记词生成公私钥对
func (b *bitcoinKeys) newKeypair() {
	//它返回一个实现了P-256（也被称为secp256r1或prime256v1）的椭圆曲线。
	//P-256是NIST（美国国家标准与技术研究院）定义的一组椭圆曲线之一，广泛用于公钥加密和数字签名。
	curve := elliptic.P256()
	var err error
	// 前四十位助记词拼接的字节数组
	buf := bytes.NewReader(b.jointSpeed())
	// 通过椭圆曲线和助记词byte数组生成私钥
	b.PrivateKey, err = ecdsa.GenerateKey(curve, buf)
	if err != nil {
		return
	}
	b.PublicKey = append(b.PrivateKey.PublicKey.X.Bytes(), b.PrivateKey.PublicKey.Y.Bytes()...)
}

// 将助记词拼接成字节数组，并截取前40位
func (b bitcoinKeys) jointSpeed() []byte {
	bs := make([]byte, 0)
	for _, v := range b.MnemonicWord {
		//将助记词转换成字节数组
		bs = append(bs, []byte(v)...)
	}
	//截取前40位
	return bs[:40]
}
