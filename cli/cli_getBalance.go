package cli

import (
	"block_chain_golang/block"
	"fmt"
)

func (cli *Cli) getBalance(address string) {
	bc := block.NewBlockchain()
	balance := bc.GetBalance(address)
	fmt.Printf("地址：%s的余额为：%d\n", address, balance)
}
