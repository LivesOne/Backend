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

	logger.InitLogger(filepath.Join(appbase, logsDir))
	logger.Info("server initialize.....")

	cfgFile := filepath.Join(appbase, configFile)
	config.LoadConfig(cfgFile)
}
