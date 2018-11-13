package config

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"servlets/common"
	"servlets/token"
	"strings"
	"testing"
	"utils"
	"utils/config"
	"utils/logger"
)

func init()  {
	initialize("D:/code/maxthon/cloud/Backend/config/config.json")
}

func TestBatchCurrencyPriceHandler_Handle(t *testing.T) {
	handler := new(batchCurrencyPriceHandler)
	req := &batchCurrencyPriceRequest{
		Base: &common.BaseInfo{
			App: &common.AppInfo{
				Name: "app",
				Ver: "1.0",
			},
		},
		Param: &batchCurrencyPriceParam{
			Currency: []string{"LVTC,CNYT","EOS,CNY"},
		},
	}
	reader := strings.NewReader(utils.ToJSON(req))
	httpReq, _ := http.NewRequest("POST", "/", reader)
	res := httptest.NewRecorder()
	handler.Handle(httpReq, res)
	t.Log(res.Body)
}

func TestTransferFeeHandler_Handle(t *testing.T) {
	handler := new(transferFeeHandler)
	req := &transferFeeRequest{
		Base: &common.BaseInfo{
			App: &common.AppInfo{
				Name: "app",
				Ver: "1.0",
			},
		},
		Param: &TransferFeeParam{
			Currency: "eos",
		},
	}
	reader := strings.NewReader(utils.ToJSON(req))
	httpReq, _ := http.NewRequest("POST", "/", reader)
	res := httptest.NewRecorder()
	handler.Handle(httpReq, res)
	t.Log(res.Body)
}

func TestWithdrawalFeeHandler_Handle(t *testing.T) {
	handler := new(withdrawalFeeHandler)
	req := &withdrawalFeeRequest{
		Base: &common.BaseInfo{
			App: &common.AppInfo{
				Name: "app",
				Ver: "1.0",
			},
		},
		Param: &WithdrawalFeeParam{
			Currency: "eos",
		},
	}
	reader := strings.NewReader(utils.ToJSON(req))
	httpReq, _ := http.NewRequest("POST", "/", reader)
	res := httptest.NewRecorder()
	handler.Handle(httpReq, res)
	t.Log(res.Body)
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
	//初始化配置主文件
	config.LoadConfig(cfgPath, cfgDir)
	fmt.Println("init config over file path ", cfgPath)
	//加载用户等级相关配置
	config.LoadLevelConfig(cfgDir, config.GetConfig().UserLevelConfig)
	//加载绑定活动相关配置
	config.LoadBindActiveConfig(cfgDir, config.GetConfig().BindActive)
	//加载提币相关配置
	config.LoadWithdrawalConfig(cfgDir, config.GetConfig().WithdrawalConfig)
	//加载log配置
	logger.InitLogger(cfgDir, config.GetConfig().LogConfig)
	logger.Info("server initialize.....")

	common.RedisPoolInit()
	common.UserDbInit()
	common.AssetDbInit()
	common.ConfigDbInit()
	common.InitTxHistoryMongoDB()
	common.InitMinerRMongoDB()
	common.InitTradeMongoDB()
	common.InitContactsMongoDB()
	token.Init()
}
