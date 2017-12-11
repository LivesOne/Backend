package main

import (
	"config"
	"lvtlog"
	"path/filepath"
	"server"
	"servlets"
	"servlets/httpcfg"
	"utils"
)

func main() {

	initialize()

	httpHandlers.RegisterHandlers()
	server.StartServer()

}

func initialize() {

	const (
		// configuration file name
		configFile = "config/config.json"
		// logs directory
		logsDir = "logs"
		// logDir = ""
	)

	appbase := utils.GetAppBaseDir()

	lvtlog.InitLogger(appbase + logsDir)
	lvtlog.Info("server initialize.....")

	cfgFile := filepath.Join(appbase, configFile)
	config.LoadConfig(cfgFile)

	cfg := config.GetConfig()
	// fmt.Println(configFile, cfg)
	httpCfg.InitHTTPConfig(cfg.ServerAddr, cfg.ServerPort)

}
