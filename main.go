package main

import (
	"block_chain_golang/block"
	"block_chain_golang/database"
	"block_chain_golang/network"
	"fmt"
	log "github.com/corgi-kx/logcustom"
	"github.com/spf13/viper"
	"os"
)

func init() {
	viper.SetConfigFile("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	logPath := viper.GetString("blockchain.log_path")
	listenHost := viper.GetString("network.listen_host")
	listenPort := viper.GetString("network.listen_port")
	rendezvousString := viper.GetString("network.rendezvous_string")
	protocolID := viper.GetString("network.protocol_id")
	tokenRewardNum := viper.GetInt("blockchain.token_reward_num") //挖矿奖励代币数量
	tradePoolLength := viper.GetInt("blockchain.trade_pool_length")
	mineDifficultyValue := viper.GetInt("blockchain.mine_difficulty_value")
	chineseMnwordPath := viper.GetString("blockchain.chinese_mnemonic_path") //中文助记词

	network.TradePoolLength = tradePoolLength
	network.ListenPort = listenPort
	network.ListenHost = listenHost
	network.RendezvousString = rendezvousString
	network.ProtocolID = protocolID
	database.ListenPort = listenPort
	block.ListenPort = listenPort
	block.TokenRewardNum = tokenRewardNum
	block.TargetBits = uint(mineDifficultyValue) // 挖矿难度
	block.ChineseMnwordPath = chineseMnwordPath

	//将日志输出到指定文件
	file, err := os.OpenFile(fmt.Sprintf("%slog%s.txt", logPath, listenPort), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		log.Error(err)
	}
	log.SetOutputAll(file)
}
func main() {
	
}
