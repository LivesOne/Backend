package config

import (
	"errors"
	"fmt"
	"path/filepath"
	"utils"
)

const USER_LEVEL_NUM = 5

type UserLevelLimit struct {
	changePhone            bool
	lockAsset              bool
	transferTo             bool
	withdrawal             bool
	singleAmountMin        int64
	singleAmountMax        int64
	dailyAmountMax         int64
	dailyPrepareAccess     int
	dailyCommitAccess      int
	dailyWithdrawalQuota   int64
	monthlyWithdrawalQuota int64
}

func (limit *UserLevelLimit) ChangePhone() bool {
	return limit.changePhone
}

func (limit *UserLevelLimit) LockAsset() bool {
	return limit.lockAsset
}

func (limit *UserLevelLimit) TransferTo() bool {
	return limit.transferTo
}

func (limit *UserLevelLimit) SingleAmountMin() int64 {
	return limit.singleAmountMin
}

func (limit *UserLevelLimit) SingleAmountMax() int64 {
	return limit.singleAmountMax
}

func (limit *UserLevelLimit) DailyAmountMax() int64 {
	return limit.dailyAmountMax
}

func (limit *UserLevelLimit) DailyPrepareAccess() int {
	return limit.dailyPrepareAccess
}

func (limit *UserLevelLimit) DailyCommitAccess() int {
	return limit.dailyCommitAccess
}

func (limit *UserLevelLimit) DailyWithdrawalQuota() int64 {
	return limit.dailyWithdrawalQuota
}

func (limit *UserLevelLimit) MonthlyWithdrawalQuota() int64 {
	return limit.monthlyWithdrawalQuota
}

func (limit *UserLevelLimit) Withdrawal() bool {
	return limit.withdrawal
}

type UserLevelConfig struct {
	LimitMap map[int]UserLevelLimit
}

type UserLevelLimitInternal struct {
	ChangePhone            bool
	LockAsset              bool
	TransferTo             bool
	Withdrawal             bool
	SingleAmountMin        int64
	SingleAmountMax        int64
	DailyAmountMax         int64
	DailyPrepareAccess     int
	DailyCommitAccess      int
	DailyWithdrawalQuota   int64
	MonthlyWithdrawalQuota int64
}

type UserLevelConfigInternal struct {
	LimitMap map[int]UserLevelLimitInternal `json:"UserLevelLimit"`
}

var gUserLevelConfig UserLevelConfig
var gUserLevelLimitDefault UserLevelLimit = UserLevelLimit{
	changePhone:            false,
	lockAsset:              false,
	transferTo:             false,
	withdrawal:             false,
	singleAmountMin:        0,
	singleAmountMax:        0,
	dailyAmountMax:         0,
	dailyPrepareAccess:     0,
	dailyCommitAccess:      0,
	dailyWithdrawalQuota:   0,
	monthlyWithdrawalQuota: 0,
}

/*func GetLevelConfig() *UserLevelConfig {
	return &gUserLevelConfig
}*/

func LoadLevelConfig(dir string, cfgName string) error {
	basePath := filepath.Join(dir, cfgName)
	fmt.Println("init level config over file path ", basePath)
	var config UserLevelConfigInternal
	err := utils.ReadJSONFile(basePath, &config)
	if err != nil {
		fmt.Println("read user level limit file error ", err)
		panic(err)
	}

	if config.isValid() == false {
		err := errors.New("user level limit item not integrity")
		fmt.Println("user level limit item not integrity\n", err)
		fmt.Println("json str --- >", utils.ToJSONIndent(config))
		panic(err)
	}

	gUserLevelConfig.LimitMap = make(map[int]UserLevelLimit)
	for level, v := range config.LimitMap {
		fmt.Println(level, v)
		gUserLevelConfig.LimitMap[level] = UserLevelLimit{
			changePhone:            v.ChangePhone,
			lockAsset:              v.LockAsset,
			transferTo:             v.TransferTo,
			withdrawal:             v.Withdrawal,
			singleAmountMin:        v.SingleAmountMin,
			singleAmountMax:        v.SingleAmountMax,
			dailyAmountMax:         v.DailyAmountMax,
			dailyPrepareAccess:     v.DailyPrepareAccess,
			dailyCommitAccess:      v.DailyCommitAccess,
			dailyWithdrawalQuota:   v.DailyWithdrawalQuota,
			monthlyWithdrawalQuota: v.MonthlyWithdrawalQuota,
		}
	}
	return nil
}

func GetLimitByLevel(level int) *UserLevelLimit {
	if lim, ok := gUserLevelConfig.LimitMap[level]; ok {
		return &lim
	}
	return &gUserLevelLimitDefault
}

func (cfg *UserLevelConfigInternal) isValid() bool {
	if len(cfg.LimitMap) < USER_LEVEL_NUM {
		fmt.Println("level is not enough")
		return false
	}

	for level := range cfg.LimitMap {
		if level < 0 || level >= USER_LEVEL_NUM {
			return false
		}
	}

	return true
}
