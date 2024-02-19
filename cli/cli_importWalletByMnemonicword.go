package cli

import (
	"block_chain_golang/block"
	"block_chain_golang/database"
	"encoding/json"
	"fmt"
	log "github.com/corgi-kx/logcustom"
)

func (cli *Cli) importWalletByMnemonicword(mnemonicword string) {
	mnemonicwords := []string{}
	err := json.Unmarshal([]byte(mnemonicword), mnemonicwords)
	if err != nil {
		log.Error("json err:", err)
	}

	bd := database.New()
	wallets := block.NewWallets(bd)
	address, privKey, mnemonicWord := wallets.GenerateWallet(bd, block.CreateBitcoinKeysByMnemonicWord, mnemonicwords)
	fmt.Println("助记词：", mnemonicWord)
	fmt.Println("私钥：", privKey)
	fmt.Println("地址：", address)
}
