package cli

import (
	"block_chain_golang/block"
	"block_chain_golang/network"
	"fmt"
)

func (cli Cli) transfer(from, to, amount string) {
	blc := block.NewBlockchain()
	blc.CreateTransaction(from, to, amount, network.Send{})
	fmt.Println("已执行转账命令")
}
