package main

import (
	"fmt"
	"os"
	"path/filepath"
	"server"
	"servlets"
	"servlets/common"
	"servlets/log_cleaner"
	"utils"
	"utils/config"
	"utils/logger"
)

func main() {
	cfgPath := ""
	if len(os.Args) > 1 {
		cfgPath = os.Args[1]
	}

	initialize(cfgPath)
	servlets.Init()
	servlets.RegisterHandlers()
	log_cleaner.StartJob()
	server.Start(config.GetConfig().ServerAddr)

}

func initialize(cfgPath string) {

	const (
		// configuration file name
		configFile = "../config/config.json"

		configDir = "../config"
	)

	fmt.Println("-----" + utils.Sha256(utils.Sha256("111111") + "180753870"))

	appbase := utils.GetAppBaseDir()

	cfgDir := filepath.Join(appbase, configDir)

	if len(cfgPath) == 0 {
		cfgPath = filepath.Join(appbase, configFile)
	} else {
		cfgDir = filepath.Join(cfgPath, "../")
	}
	//初始化配置主文件
	config.LoadConfig(cfgPath, cfgDir)
	fmt.Println("init config over file path ", cfgPath)
	//加载用户等级相关配置
	config.LoadLevelConfig(cfgDir, config.GetConfig().UserLevelConfig)
	//加载绑定活动相关配置
	config.LoadBindActiveConfig(cfgDir, config.GetConfig().BindActive)
	//加载提币相关配置
	config.LoadWithdrawalConfig(cfgDir, config.GetConfig().WithdrawalConfig)
	//加载log配置
	logger.InitLogger(cfgDir, config.GetConfig().LogConfig)
	logger.Info("server initialize.....")

	go func() {
		common.ListenTxhistoryQueue()
	}()
	go func() {
		common.PushTxHistoryByTimer()
	}()
}
