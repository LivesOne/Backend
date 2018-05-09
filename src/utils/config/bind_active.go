package config

import (
	"path/filepath"
	"fmt"
	"utils"
	"errors"
)



type BindActive struct {
	RegisterTimeActive int64
	BindTimeActiveStart int64
	BindTimeActiveEnd int64
	BindWXActiveHashRate int
	BindTGActiveHashRate int
	HashRateActiveMonth int

}

var config *BindActive

/*func GetLevelConfig() *UserLevelConfig {
	return &gUserLevelConfig
}*/

func LoadBindActiveConfig(dir string, cfgName string) error {
	basePath := filepath.Join(dir, cfgName)
	fmt.Println("init level config over file path ", basePath)
	config = new(BindActive)
	err := utils.ReadJSONFile(basePath, config)
	if err != nil {
		fmt.Println("read user level limit file error ", err)
		panic(err)
	}

	if !config.isValid(){
		err := errors.New("user level limit item not integrity")
		fmt.Println("user level limit item not integrity\n", err)
		fmt.Println("json str --- >", utils.ToJSONIndent(config))
		panic(err)
	}

	return nil
}


func (cfg *BindActive) isValid() bool {
	return cfg.RegisterTimeActive >0 &&
		cfg.BindTGActiveHashRate >0 &&
		cfg.BindTimeActiveEnd >0 &&
		cfg.BindTimeActiveStart >0 &&
		cfg.HashRateActiveMonth >0
}

func GetBindActive()*BindActive{
	return config
}