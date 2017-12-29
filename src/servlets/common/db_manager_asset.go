package common

import (
	_ "fmt"
	"utils/config"
	_ "utils/config"
	"utils/db_factory"
	"utils/logger"


	_ "github.com/go-sql-driver/mysql"
)

//var gDbUser *sql.DB
var gDBAsset *db_factory.DBPool
func AssetDbInit() error {

	db_config_asset := config.GetConfig().User
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

