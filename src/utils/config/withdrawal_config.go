package config

import (
	"errors"
	"fmt"
	"path/filepath"
	"utils"
)

type WithdrawalConfig struct {
	WithdrawalAcceptAccount        int64
	FeeAcceptAccount               int64
	WithdrawalCardEthAcceptAccount int64
	WithdrawalCardEthUnitPrice     float64
}

var withdrawalConfig *WithdrawalConfig

func LoadWithdrawalConfig(dir string, cfgName string) error {
	basePath := filepath.Join(dir, cfgName)
	fmt.Println("init withdrawal config over file path ", basePath)
	withdrawalConfig = new(WithdrawalConfig)
	err := utils.ReadJSONFile(basePath, withdrawalConfig)
	if err != nil {
		fmt.Println("read withdrawal config limit file error ", err)
		panic(err)
	}

	if !withdrawalConfig.isValid() {
		err := errors.New("withdrawal config item not integrity")
		fmt.Println("withdrawal config item not integrity\n", err)
		fmt.Println("json str --- >", utils.ToJSONIndent(withdrawalConfig))
		panic(err)
	}

	return nil
}

func (cfg *WithdrawalConfig) isValid() bool {
	return cfg.WithdrawalAcceptAccount > 0 &&
		cfg.FeeAcceptAccount > 0 &&
		cfg.WithdrawalCardEthUnitPrice > 0 &&
		cfg.WithdrawalCardEthAcceptAccount > 0
}

func GetWithdrawalConfig() *WithdrawalConfig {
	return withdrawalConfig
}
