package config

import (
	"errors"
	"io/ioutil"
	"path/filepath"
	"utils"
	"utils/logger"
)

type DBConfig struct {
	// mysql config
	MaxConn    int
	DBHost     string
	DBUser     string
	DBUserPwd  string
	DBDatabase string
}

type RedisConfig struct {
	MaxConn    int
	RedisAddr string //"[ip]:port"
	RedisAuth string
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
	TxHistory DBConfig
	// redis的参数
	Redis RedisConfig

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

	err := utils.ReadJSONFile(cfgFilename, &gConfig)
	if err != nil {
		return err
	}

	if gConfig.isValid() == false {
		logger.Info("configuration item not integrity\n", utils.ToJSONIndent(gConfig))
		return errors.New("configuration item not integrity")
	}
	// logger.Info("load configuration success")

	return nil
}

// GetConfig get the config data
func GetConfig() *Configuration {

	return &gConfig
}

// GetPrivateKey reads the private key from the key file
// @param ver:  version of server public key
func GetPrivateKey(ver int) ([]byte, error) {
	// right now, we ONLY have the pri/pub key pair of version 1
	if ver != 1 {
		logger.Info("GetPrivateKey by public key version: no corresponding private key, version#", ver)
		return nil, errors.New("param error")
	}

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

	return gPrivKeyContent, nil
}

// GetPrivateKeyFilename returns the rsa private key file name
func GetPrivateKeyFilename() string {
	return filepath.Join(utils.GetAppBaseDir(), "../config", gConfig.PrivKey)
}

func (db *DBConfig) isValid() bool {

	return db.MaxConn > 0 &&
		len(db.DBHost) > 0 &&
		len(db.DBUser) > 0 &&
		len(db.DBUserPwd) > 0 &&
		len(db.DBDatabase) > 0

}

func (db *RedisConfig) isValid() bool {

	return db.MaxConn > 0 &&
		len(db.RedisAddr) > 0 &&
		len(db.RedisAuth) > 0
}


func (cfg *Configuration) isValid() bool {

	return len(cfg.ServerAddr) > 0 &&
		len(cfg.PrivKey) > 0 &&
		cfg.User.isValid() &&
		cfg.Asset.isValid() &&
		cfg.Redis.isValid() &&
		len(cfg.SmsSvrAddr) > 0 &&
		len(cfg.MailSvrAddr) > 0 &&
		len(cfg.ImgSvrAddr) > 0
}
