package block

/*
	utxo数据库创建的意义在于,不会每次进行转帐时遍历整个区块链,
	而是去utxo数据库查找未消费的交易输出,这样会大大降低性能问题
*/

type UTXOHandle struct {
	BC *blockchain
}

// 重置UTXO数据库
func (u *UTXOHandle) ResetUTXODataBase() {
	// 先查找全部未花费UTXO

}
