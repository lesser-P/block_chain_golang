package cli

import (
	"fmt"
	log "github.com/corgi-kx/logcustom"
	"strconv"
	"strings"
)

type Cli struct {
}

// 打印帮助提示
func printUsage() {
	fmt.Println("----------------------------------------------------------------------------- ")
	fmt.Println("Usage:")
	fmt.Println("\thelp                                              打印命令行说明")
	fmt.Println("\tgenesis  -a DATA  -v DATA                         生成创世区块")
	fmt.Println("\tsetRewardAddr -a DATA                             设置挖矿奖励地址")
	fmt.Println("\tgenerateWallet                                    创建新钱包")
	fmt.Println("\timportMnword -m DATA                              根据助记词导入钱包")
	fmt.Println("\tprintAllWallets                                   查看本地存在的钱包信息")
	fmt.Println("\tprintAllAddr                                      查看本地存在的地址信息")
	fmt.Println("\tgetBalance  -a DATA                               查看用户余额")
	fmt.Println("\ttransfer -from DATA -to DATA -amount DATA         进行转账操作")
	fmt.Println("\tprintAllBlock                                     查看所有区块信息")
	fmt.Println("\tresetUTXODB                                       遍历区块数据，重置UTXO数据库")
	fmt.Println("------------------------------------------------------------------------------")
}

func New() *Cli {
	return &Cli{}
}

func (cli *Cli) Run() {
	printUsage()
}

// 用户输入命令解析
func (cli *Cli) userCmdHandle(data string) {
	// 去除命令前后空格
	strings.TrimSpace(data)
	var cmd string
	var context string
	// 第一个空格前是命令后面是命令的参数
	if strings.Contains(data, " ") {
		cmd = data[:strings.Index(data, " ")]
		context = data[strings.Index(data, " ")+1:]
	} else {
		cmd = data
	}

	switch cmd {
	case "help":
		printUsage()
	case "genesis":
		address := getSpecifiedContent(data, "-a", "-v")
		value := getSpecifiedContent(data, "-v", "")
		v, err := strconv.Atoi(value)
		if err != nil {
			log.Fatal(err)
		}
		cli.genesis(address, v)
	case "generateWallet":
		cli.generateWallet()
	case "setRewardAddr":
		address := getSpecifiedContent(data, "-a", "")
		cli.setRewardAddress(address)
	case "importMnword":
		// 导入助记词获得公私钥
		mnemonicword := getSpecifiedContent(data, "-m", "")
		cli.importWalletByMnemonicword(mnemonicword)
	case "printAllAddr":
		//打印所有钱包地址
		cli.printAllAddress()
	case "printAllWallets":
		//打印所有钱包明细
		cli.printAllWallets()
	case "printAllBlock":
		//打印全部区块详情
		cli.printAllBlock()
	case "getBalance":
		address := getSpecifiedContent(data, "-a", "")
		cli.getBalance(address)
	case "resetUTXODB":
		cli.resetUTXODB()
	case "transfer":
		//发送交易
		//截取fromaddress
		fromString := (context[strings.Index(context, "-from")+len("-from") : strings.Index(context, "-to")])
		toString := strings.TrimSpace(context[strings.Index(context, "-to")+len("-to") : strings.Index(context, "-amount")])
		amountString := strings.TrimSpace(context[strings.Index(context, "-amount")+len("-amount"):])
		cli.transfer(fromString, toString, amountString)
	default:
		fmt.Println("无此命令")
		printUsage()
	}
}

// 返回data字符串中，标签为tag的内容
func getSpecifiedContent(data, tag, end string) string {
	if end != "" {
		// tag-end
		return strings.TrimSpace(data[strings.Index(data, tag)+len(tag) : strings.Index(data, end)])
	}
	// tag-结尾
	return strings.TrimSpace(data[strings.Index(data, tag)+len(tag):])
}
