package block

// UTXO输出
type TxOutput struct {
	Value         int    //交易金额
	PublicKeyHash []byte //公钥哈希
}
