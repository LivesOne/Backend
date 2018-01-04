package common

import (
	_ "fmt"
	_ "github.com/go-sql-driver/mysql"
	"utils"
	"utils/config"
	_ "utils/config"
	"utils/db_factory"
	"utils/logger"
)

//var gDbUser *sql.DB
var gDBAsset *db_factory.DBPool

func AssetDbInit() error {

	db_config_asset := config.GetConfig().Asset
	facConfig_asset := db_factory.Config{
		Host:        db_config_asset.DBHost,
		UserName:    db_config_asset.DBUser,
		Password:    db_config_asset.DBUserPwd,
		Database:    db_config_asset.DBDatabase,
		MaxConn:     10,
		MaxIdleConn: 1,
	}
	gDBAsset = db_factory.NewDataSource(facConfig_asset)
	if gDBAsset.IsConn() {
		logger.Debug("connection database successful")
	} else {
		logger.Fatal(gDBAsset.Err())
	}

	return nil
}

func QueryReward(uid int64) *Reward {
	row, err := gDBAsset.QueryRow("select total,yesterday,lastmodify from user_reward where uid = ?", uid)
	if err != nil {
		logger.Error("query db error ", err.Error())
	}
	return &Reward{
		Total:      utils.Str2Int64(row["total"]),
		Yesterday:  utils.Str2Int64(row["yesterday"]),
		Lastmodify: utils.Str2Int64(row["lastmodify"]),
		Uid:        uid,
	}
}
