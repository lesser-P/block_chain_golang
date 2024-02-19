package cli

import "block_chain_golang/network"

func (cli Cli) startNode() {
	network.StartNode(cli)
}
