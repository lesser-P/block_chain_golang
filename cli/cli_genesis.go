package cli

import (
	"block_chain_golang/block"
	"block_chain_golang/network"
)

// 创建创世区块。两个参数为接收创世区块奖励的地址和创世区块奖励数量
func (cli *Cli) genesis(address string, value int) {
	bc := block.NewBlockchain()
	bc.CreateGenesisTransaction(address, value, network.Send{})
}
