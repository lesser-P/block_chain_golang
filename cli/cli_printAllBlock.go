package cli

import "block_chain_golang/block"

func (cli *Cli) printAllBlock() {
	bc := block.NewBlockchain()
	bc.PrintAllBlockInfo()
}
