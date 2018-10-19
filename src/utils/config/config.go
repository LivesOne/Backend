package config

import (
	"errors"
	"fmt"
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

type MongoConfig struct {
	// mysql config
	MaxConn    int
	DBHost     string
	DBUser     string
	DBUserPwd  string
	DBDatabase string
}

type RedisConfig struct {
	MaxConn   int
	RedisAddr string //"[ip]:port"
	RedisAuth string
	DBIndex   int
}

type captcha struct {
	Url       string
	Id        string
	SecretId  string
	SecretKey string
}

type TransferLimit struct {
	SingleAmountMin    int64
	SingleAmountMax    int64
	DailyAmountMax     int64
	DailyPrepareAccess int
	DailyCommitAccess  int
}

type LoginPwdErrCntLimit struct {
	Number int
	Min    int
}

type WXAuth struct {
	Url    string
	Appid  string
	Secret string
}

type ReChargeAddr struct {
	Currency string
	Address  string
}

// Configuration holds all config data
type Configuration struct {
	ServerAddr string //"[ip]:port"

	// server side private key file name
	PrivKey string

	// account db config
	User DBConfig
	// asset db config
	Asset        DBConfig
	TxHistory    MongoConfig
	NewTxHistory MongoConfig
	Miner        MongoConfig
	Trade        MongoConfig
	Config       MongoConfig
	// redis的参数
	Redis RedisConfig
	//密码错误登陆限制
	LoginPwdErrCntLimit []LoginPwdErrCntLimit
	TransferLimit       map[int]TransferLimit
	AppIDs              []int // app IDs read from configuration file

	// 短信验证网关相关
	SmsSvrAddr string
	// 邮件验证网关相关
	MailSvrAddr string
	// 图像验证网关相关
	ImgSvrAddr string
	// 短信上行接口地址
	SmsUpValidateSvrAddr string
	// log相关
	LogConfig                     string
	UserLevelConfig               string
	Captcha                       captcha
	MaxActivityRewardValue        int
	CautionMoneyIds               []int64
	PenaltyMoneyAccountUid        int64
	WXAuth                        WXAuth
	BindActive                    string
	AuthTelegramUrl               string
	WithdrawalConfig              string
	LvtcHashrateScale             int
	Lvt2LvtcSystemAccountUid      int64
	Lvt2LvtcDelaySystemAccountUid int64
	TransFeeAccountUid            int64
	ReChargeAddress               []ReChargeAddr
}

// configuration data
var gConfig Configuration

// RSA private key content read from the private key file
var gPrivKeyContent []byte

var cfgDir string

// LoadConfig load the configuration from the configuration file
func LoadConfig(cfgFilename string, cd string) error {

	err := utils.ReadJSONFile(cfgFilename, &gConfig)
	if err != nil {
		fmt.Println("read json file error ", err)
		panic(err)
	}

	if gConfig.isValid() == false {
		err := errors.New("configuration item not integrity")
		fmt.Println("configuration item not integrity\n", err)
		fmt.Println("json str --- >", utils.ToJSONIndent(gConfig))
		panic(err)
		//return errors.New("configuration item not integrity")
	}
	//logger.Info(gConfig.AppIDs)
	//gConfig.appsMap = make(map[string]bool)
	//for _, appid := range gConfig.AppIDs {
	//	gConfig.appsMap[appid] = true
	//}
	//gConfig.AppIDs = nil // release it
	// logger.Info("load configuration success, is app id valid:", gConfig.IsAppIDValid("maxthon"))
	// logger.Info("configuration item not integrity\n", utils.ToJSONIndent(gConfig))
	cfgDir = cd
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
		logger.Info("PrivateKey path ", filename)
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
	return filepath.Join(cfgDir, gConfig.PrivKey)
}

func (db *DBConfig) isValid() bool {

	return db.MaxConn > 0 &&
		len(db.DBHost) > 0 &&
		len(db.DBUser) > 0 &&
		len(db.DBUserPwd) > 0 &&
		len(db.DBDatabase) > 0

}

func (db *MongoConfig) isValid() bool {

	return db.MaxConn > 0 &&
		len(db.DBHost) > 0 &&
		len(db.DBDatabase) > 0

}

func (db *RedisConfig) isValid() bool {
	return db.MaxConn > 0 && len(db.RedisAddr) > 0
}

func (cfg *Configuration) isValid() bool {

	return len(cfg.ServerAddr) > 0 &&
		len(cfg.PrivKey) > 0 &&
		cfg.User.isValid() &&
		cfg.Asset.isValid() &&
		cfg.Redis.isValid() &&
		cfg.TxHistory.isValid() &&
		//len(cfg.AppIDs) > 0 &&
		len(cfg.SmsSvrAddr) > 0 &&
		len(cfg.MailSvrAddr) > 0 &&
		len(cfg.ImgSvrAddr) > 0
}

func (cfg *Configuration) CautionMoneyIdsExist(uid int64) bool {
	if len(cfg.CautionMoneyIds) > 0 {
		for _, v := range cfg.CautionMoneyIds {
			if uid == v {
				return true
			}
		}
	}
	return false
}
func IsAppIDValid(appid int) bool {
	logger.Info("app_id in ", appid, "curr app_id ", gConfig.AppIDs)
	for _, v := range gConfig.AppIDs {
		if v == appid {
			return true
		}
	}
	return false
}
