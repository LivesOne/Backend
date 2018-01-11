package main

import (
	"path/filepath"
	"server"
	"servlets"
	"utils"
	"utils/config"
	"utils/logger"
)

func main() {

	initialize()
	servlets.Init()
	servlets.RegisterHandlers()
	server.Start(config.GetConfig().ServerAddr)

}

func initialize() {

	const (
		// configuration file name
		configFile = "../config/config.json"
		// logs directory
		logsDir = "logs"
		// logDir = ""
	)

	appbase := utils.GetAppBaseDir()



	cfgFile := filepath.Join(appbase, configFile)
	config.LoadConfig(cfgFile)

	logger.InitLogger(config.GetConfig().LogDir,config.GetConfig().LoggerLevel)
	logger.Info("server initialize.....")
}
