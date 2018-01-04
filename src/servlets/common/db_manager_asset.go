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

const  (
	CONV_LVT = 10000*10000
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
	row, err := gDBAsset.QueryRow("select total,lastday,lastmodify from user_reward where uid = ?", uid)
	if err != nil {
		logger.Error("query db error ", err.Error())
	}
	return &Reward{
		Total:      utils.Str2Int64(row["total"]),
		Yesterday:  utils.Str2Int64(row["lastday"]),
		Lastmodify: utils.Str2Int64(row["lastmodify"]),
		Uid:        uid,
	}
}


func QueryBalance(uid int64)int64{
	row, err := gDBAsset.QueryRow("select balance from user_asset where uid = ?", uid)
	if err != nil {
		logger.Error("query db error ", err.Error())
	}

	if row != nil {
		return utils.Str2Int64(row["balance"])
	}
	return 0
}

func TransAccountLvt(from,to,value int64)(bool){
	tx,err := gDbUser.Begin()
	if err!=nil {
		logger.Error("db pool begin error ",err.Error())
		return false
	}
	tx.Exec("select * from user_asset where uid in (?,?)",from,to)

	//查询转出账户余额是否满足需要
	var balance int64
	row := tx.QueryRow("select balance from user_asset where uid  = ?",from)
	row.Scan(&balance)

	if balance < value {
		tx.Rollback()
		return false
	}


	_,err1 := tx.Exec("update user_asset set balance = balance - ? where uid = ?",value,from)
	if err1 != nil {
		logger.Error("sql error ",err1.Error())
		tx.Rollback()
		return false
	}
	_,err2 := tx.Exec("update user_asset set balance = balance + ? where uid = ?",value,to)
	if err2 != nil {
		logger.Error("sql error ",err2.Error())
		tx.Rollback()
		return false
	}
	tx.Commit()
	return true
}