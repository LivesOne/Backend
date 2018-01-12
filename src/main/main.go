package main

import (
	"path/filepath"
	"server"
	"servlets"
	"utils"
	"utils/config"
	"os"
	"fmt"
	"utils/logger"
)

func main() {
	cfgPath := ""
	if len(os.Args) > 1{
		cfgPath = os.Args[1]
	}
	initialize(cfgPath)
	servlets.Init()
	servlets.RegisterHandlers()
	server.Start(config.GetConfig().ServerAddr)

}

func initialize(cfgPath string) {

	const (
		// configuration file name
		configFile = "../config/config.json"
		// logs directory
		logsDir = "logs"
		// logDir = ""
	)





	appbase := utils.GetAppBaseDir()
	if len(cfgPath) == 0 {
		cfgPath = filepath.Join(appbase, configFile)
	}
	//cfgFile := filepath.Join(appbase, configFile)

	fmt.Println("init config file path ",cfgPath)
	config.LoadConfig(cfgPath)

	logger.InitLogger(config.GetConfig().LogConfigPath,appbase)
	logger.Info("server initialize.....")
}
