package block

// UTXO输入
type TxInput struct {
	TxHash    []byte //交易哈希
	Index     int
	Signature []byte //签名
	PublicKey []byte //公钥
}
