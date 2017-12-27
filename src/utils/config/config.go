package config

import (
	"io/ioutil"
	"path/filepath"
	"utils"
	"utils/logger"
)

type DBConfig struct {
	// mysql config
	DBHost     string
	DBUser     string
	DBUserPwd  string
	DBDatabase string
}

// Configuration holds all config data
type Configuration struct {
	ServerAddr string //"[ip]:port"

	// server side private key file name
	PrivKey string

	// account db config
	User DBConfig
	// asset db config
	Asset DBConfig

	// redis的参数
	RedisAddr string //"[ip]:port"
	RedisAuth string

	// 短信验证网关相关
	SmsSvrAddr string
	// 邮件验证网关相关
	MailSvrAddr string
	// 图像验证网关相关
	ImgSvrAddr string
	// log相关
}

// configuration data
var gConfig Configuration

// RSA private key content read from the private key file
var gPrivKeyContent []byte

// LoadConfig load the configuration from the configuration file
func LoadConfig(cfgFilename string) error {

	return utils.ReadJSONFile(cfgFilename, &gConfig)

}

// GetConfig get the config data
func GetConfig() *Configuration {

	return &gConfig
}

// GetPrivateKey reads the private key from the key file
func GetPrivateKey() []byte {
	if (gPrivKeyContent == nil) || (len(gPrivKeyContent) < 1) {

		filename := GetPrivateKeyFilename()
		// fmt.Println("private key file:", filename, "ddd:", gConfig.PrivKey)
		var err error
		gPrivKeyContent, err = ioutil.ReadFile(filename)
		if err != nil {
			logger.Info("load private key failed:", filename, err)
			gPrivKeyContent = nil
		}
	}

	return gPrivKeyContent
}

// GetPrivateKeyFilename returns the rsa private key file name
func GetPrivateKeyFilename() string {
	return filepath.Join(utils.GetAppBaseDir(), "../config", gConfig.PrivKey)
}
