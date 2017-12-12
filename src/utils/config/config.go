package config

import (
	"utils"
)

// Configuration holds all config data
type Configuration struct {

	// 服务bind地址
	ServerAddr string

	// 服务listen端口
	ServerPort int

	// mysql config
	DBHost     string
	DBUser     string
	DBUserPwd  string
	DBDatabase string

	// server side private key
	PrivKey string

	// redis的参数

	// 短信验证网关相关

	// 邮件验证网关相关

	// 图像验证网关相关

	// log相关
}

// configuration data
var g_config Configuration

// LoadConfig load the configuration from the configuration file
func LoadConfig(cfgFilename string) error {

	return utils.ReadJSONFile(cfgFilename, &g_config)

}

// GetConfig get the config data
func GetConfig() *Configuration {

	return &g_config
}
