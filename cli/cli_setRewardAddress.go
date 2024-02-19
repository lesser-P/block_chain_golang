package cli

import (
	"block_chain_golang/block"
	"fmt"
)

func (cli *Cli) setRewardAddress(address string) {
	bc := block.NewBlockchain()
	bc.SetRewardAddress(address)
	fmt.Printf("已设置地址%s为挖矿奖励地址！\n", address)
}
