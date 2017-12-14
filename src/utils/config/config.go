package config

import (
	"utils"
)

// Configuration holds all config data
type Configuration struct {
	ServerAddr string //"[ip]:port"

	// mysql config
	DBHost     string
	DBUser     string
	DBUserPwd  string
	DBDatabase string

	// server side private key
	PrivKey string

	// redis的参数
	RedisAddr string //"[ip]:port"

	// 短信验证网关相关

	// 邮件验证网关相关

	// 图像验证网关相关

	// log相关
}

// configuration data
var gConfig Configuration

// LoadConfig load the configuration from the configuration file
func LoadConfig(cfgFilename string) error {

	return utils.ReadJSONFile(cfgFilename, &gConfig)

}

// GetConfig get the config data
func GetConfig() *Configuration {

	return &gConfig
}
