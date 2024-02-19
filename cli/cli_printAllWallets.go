package cli

import (
	"block_chain_golang/block"
	"block_chain_golang/database"
	"fmt"
)

func (cli *Cli) printAllWallets() {
	bd := database.New()
	wallets := block.NewWallets(bd)
	if len(wallets.Wallets) == 0 {
		fmt.Println("当前节点没有生成或导入的钱包信息！")
		return
	}

	fmt.Println("已生成的钱包信息")
	fmt.Println("==================================================================")
	for k, v := range wallets.Wallets {
		fmt.Println("地址：", k)
		fmt.Printf("公钥：%x\n", v.PublicKey)
		fmt.Println("私钥：", v.GetPrivateKey())
		fmt.Println("助记词：", v.MnemonicWord)
		fmt.Println("===================================================================")
	}
}
