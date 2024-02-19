package cli

import (
	"block_chain_golang/block"
	"block_chain_golang/database"
	"fmt"
)

// 生成钱包
func (cli *Cli) generateWallet() {
	bd := database.New()
	wallets := block.NewWallets(bd)
	address, privKey, mnemonicWord := wallets.GenerateWallet(bd, block.NewBitcoinKeys, []string{})
	fmt.Println("助记词：", mnemonicWord)
	fmt.Println("私钥：", privKey)
	fmt.Println("地址：", address)
}
