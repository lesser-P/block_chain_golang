package cli

import (
	"block_chain_golang/block"
	"fmt"
)

func (cli *Cli) resetUTXODB() {
	bc := block.NewBlockchain()
	utxoHandle := block.UTXOHandle{bc}
	utxoHandle.ResetUTXODataBase()
	fmt.Println("已重置UTXO数据库")
}
