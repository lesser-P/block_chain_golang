package block

import "block_chain_golang/network"

type Sender interface {
	SendVersionToPeers(height int)
	SendTransToPeers(tss []network.Transaction)
}
