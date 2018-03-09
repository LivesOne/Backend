package main

import (
	"fmt"
	"os"
	"path/filepath"
	"server"
	"servlets"
	"utils"
	"utils/config"
	"utils/logger"
	"servlets/log_cleaner"
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

	appbase := utils.GetAppBaseDir()

	cfgDir := filepath.Join(appbase, configDir)

	if len(cfgPath) == 0 {
		cfgPath = filepath.Join(appbase, configFile)
	} else {
		cfgDir = filepath.Join(cfgPath, "../")
	}


	config.LoadConfig(cfgPath, cfgDir)
	fmt.Println("init config over file path ", cfgPath)
	logger.InitLogger(cfgDir, config.GetConfig().LogConfig)
	logger.Info("server initialize.....")
}
