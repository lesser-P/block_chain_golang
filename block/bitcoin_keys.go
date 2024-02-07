package block

import (
	"block_chain_golang/util"
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"github.com/cloudflare/cfssl/scan/crypto/sha256"
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

// 判断是否是有效的比特币地址
func IsVailBitcoinAddress(address string) bool {
	addressBytes := []byte(address)
	//对地址字节数组进行加密
	fullhash := util.Base58Decode(addressBytes)
	if len(fullhash) != 25 {
		return false
	}
	prefixHash := fullhash[:len(fullhash)-checkSum]
	tailHash := fullhash[len(fullhash)-checkSum:]
	tailHash2 := checkSumHash(prefixHash)
	// 判断是否相等
	if bytes.Compare(tailHash, tailHash2[:]) == 0 {
		return true
	}
	return false
}

// 生成公钥哈希
func generatePublicKeyHash(publicKey []byte) []byte {
	sha256PubKey := sha256.Sum256(publicKey)
	r := util.NewRipemd160()
	r.Reset()
	r.Write(sha256PubKey[:])
	ripPubKey := r.Sum(nil)
	return ripPubKey
	return nil
}

func checkSumHash(versionPublickeyHash []byte) []byte {
	sum256 := sha256.Sum256(versionPublickeyHash)
	versionPublickeyHash1 := sha256.Sum256(sum256[:])
	tailHash := versionPublickeyHash1[:checkSum]
	return tailHash
}

func getPublicKeyHashFromAddress(address string) []byte {
	addressByte := []byte(address)
	fullHash := util.Base58Decode(addressByte)
	publicKeyHash := fullHash[1 : len(fullHash)-checkSum]
	return publicKeyHash
}

// 使用私钥进行数字签名
func ellipticCurveSign(privKey *ecdsa.PrivateKey, hash []byte) []byte {
	r, s, err := ecdsa.Sign(rand.Reader, privKey, hash)
	if err != nil {
		log.Panic("EllipticCurveSign:", err)
	}
	signature := append(r.Bytes(), s.Bytes()...)
	return signature
}

// 通过公钥信息获得地址
func GetAddressFromPublicKey(publickey []byte) string {
	if publickey == nil {
		return ""
	}
	b := bitcoinKeys{PublicKey: publickey}
	return string(b.getAddress())
}

// 使用公钥进行签名验证
func ellipticCurveVerify(pubKey []byte, signature []byte, hash []byte) bool {
	//拆分签名的到r，s
	r := big.Int{}
	s := big.Int{}
	sigLen := len(signature)
	//前一半
	r.SetBytes(signature[:(sigLen / 2)])
	//后一半
	s.SetBytes(signature[(sigLen / 2):])

	//拆分公钥
	x := big.Int{}
	y := big.Int{}
	keyLen := len(pubKey)
	x.SetBytes(pubKey[:(keyLen / 2)])
	y.SetBytes(pubKey[(keyLen / 2):])

	curve := elliptic.P256()
	// 字节转publickey
	rawPubKey := ecdsa.PublicKey{curve, &x, &y}
	// 传入公钥，要验证的信息，以及签名
	if ecdsa.Verify(&rawPubKey, hash, &r, &s) == false {
		return false
	}
	return true
}

func (bk *bitcoinKeys) getAddress() []byte {
	ripPubKey := generatePublicKeyHash(bk.PublicKey)
	versionPublickeyHash := append([]byte(version), ripPubKey[:]...)
	// 取最后四个字节的值
	tailHash := checkSumHash(versionPublickeyHash)
	// 拼接最终hash versionPublicKeyHash + checkSumHash
	finalHash := append(versionPublickeyHash, tailHash...)
	address := util.Base58Encode(finalHash)
	return address
}
