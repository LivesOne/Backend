package config

import (
	"path/filepath"
	"fmt"
	"utils"
	"errors"
)

const USER_LEVEL_NUM = 5

type UserLevelLimit struct {
	ChangePhone			bool
	LockAsset			bool
	TransferTo			bool
	SingleAmountMin		int64
	SingleAmountMax		int64
	DailyAmountMax		int64
	DailyPrepareAccess	int
	DailyCommitAccess	int
}

type UserLevelConfig struct {
	UserLevelLimit map[int]UserLevelLimit
}

var gUserLevelConfig UserLevelConfig
var gUserLevelLimitDefault UserLevelLimit = UserLevelLimit {
	ChangePhone:false,
	LockAsset:false,
	TransferTo:false,
	SingleAmountMin:0,
	SingleAmountMax:0,
	DailyAmountMax:0,
	DailyPrepareAccess:0,
	DailyCommitAccess:0,
}

/*func GetLevelConfig() *UserLevelConfig {
	return &gUserLevelConfig
}*/

func LoadLevelConfig(dir string, cfgName string) error {
	basePath := filepath.Join(dir, cfgName)
	fmt.Println("init level config over file path ", basePath)
	err := utils.ReadJSONFile(basePath, &gUserLevelConfig)
	if err != nil {
		fmt.Println("read user level limit file error ", err)
		panic(err)
	}

	if gUserLevelConfig.isValid() == false {
		err := errors.New("user level limit item not integrity")
		fmt.Println("user level limit item not integrity\n", err)
		fmt.Println("json str --- >", utils.ToJSONIndent(gUserLevelConfig))
		panic(err)
	}

	return nil
}

func GetLimitByLevel(level int) *UserLevelLimit {
	if lim, ok := gUserLevelConfig.UserLevelLimit[level]; ok {
		return &lim
	}
	return &gUserLevelLimitDefault
}

func (cfg *UserLevelConfig) isValid() bool {
	if len(cfg.UserLevelLimit) < USER_LEVEL_NUM {
		fmt.Println("level is not enough");
		return false
	}

	for level, v := range cfg.UserLevelLimit {
		fmt.Println(level, v)
		if level < 0 || level >= USER_LEVEL_NUM {
			return false
		}
	}

	return true
}
