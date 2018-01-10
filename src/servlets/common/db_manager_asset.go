package common

import (
	_ "fmt"
	_ "github.com/go-sql-driver/mysql"
	"utils"
	"utils/config"
	_ "utils/config"
	"utils/db_factory"
	"utils/logger"
	"database/sql"
	"servlets/constants"
)

const  (
	CONV_LVT = 10000*10000
)
//var gDBAsset *sql.DB
var gDBAsset *db_factory.DBPool

func AssetDbInit() error {

	db_config_asset := config.GetConfig().Asset
	facConfig_asset := db_factory.Config{
		Host:        db_config_asset.DBHost,
		UserName:    db_config_asset.DBUser,
		Password:    db_config_asset.DBUserPwd,
		Database:    db_config_asset.DBDatabase,
		MaxConn:     db_config_asset.MaxConn,
		MaxIdleConn: db_config_asset.MaxConn,
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



func TransAccountLvt(txid,from,to,value int64)(bool,int){
	//检测资产初始化情况
	//from 的资产如果没有初始化，初始化并返回false--》 上层检测到false会返回余额不足
	if !CheckAndInitAsset(from) {
		return false,constants.TRANS_ERR_INSUFFICIENT_BALANCE
	}
	ts := utils.GetTimestamp13()

	tx,err := gDBAsset.Begin()
	if err!=nil {
		logger.Error("db pool begin error ",err.Error())
		return false,constants.TRANS_ERR_SYS
	}
	tx.Exec("select * from user_asset where uid in (?,?) for update",from,to)

	//查询转出账户余额是否满足需要
	var balance int64
	row := tx.QueryRow("select balance from user_asset where uid  = ?",from)
	row.Scan(&balance)

	if balance < value {
		tx.Rollback()
		return false,constants.TRANS_ERR_INSUFFICIENT_BALANCE
	}

	_,err1 := tx.Exec("update user_asset set balance = balance - ?,lastmodify = ? where uid = ?",value,ts,from)
	if err1 != nil {
		logger.Error("sql error ",err1.Error())
		tx.Rollback()
		return false,constants.TRANS_ERR_SYS
	}
	_,err2 := tx.Exec("update user_asset set balance = balance + ?,lastmodify = ? where uid = ?",value,ts,to)
	if err2 != nil {
		logger.Error("sql error ",err2.Error())
		tx.Rollback()
		return false,constants.TRANS_ERR_SYS
	}

	//txid 写入数据库
	_,e := InsertTXID(txid,tx)

	if e != nil {
		logger.Error("sql error ",e.Error())
		tx.Rollback()
		return false,constants.TRANS_ERR_SYS
	}


	tx.Commit()

	return true,constants.TRANS_ERR_SUCC
}

func CheckAndInitAsset(uid int64)bool{
	uid,status := GetAssetByUid(uid)
	if status != constants.ASSET_STATUS_INIT {
		//初始化资产
		InsertAsset(uid)
		//初始化工资
		InsertReward(uid)
		//修改资产初始化状态
		SetAssetStatus(uid,constants.ASSET_STATUS_INIT )
		return false
	}
	return true
}

func InsertTXID(txid int64,tx *sql.Tx)(sql.Result,error) {
	res,err := tx.Exec("Insert into recent_tx_ids values (?)", txid)
	if err != nil {
		logger.Error("query error ",err.Error())
	}
	return res,err

}

func RemoveTXID(txid int64) error{
	_,err := gDBAsset.Exec("delete from recent_tx_ids where txid = ?", txid)
	if err != nil {
		logger.Error("query error ",err.Error())
	}
	return err
}

func InsertReward(uid int64) error {
	sql := "insert ignore into user_reward (uid,total,lastday,lastmodify) values (?,?,?,?) "
	_,err := gDBAsset.Exec(sql, uid, 0, 0, 0)
	return err
}

func InsertAsset(uid int64)error {
	sql := "insert ignore into user_asset (uid,balance,lastmodify) values (?,?,?) "
	_,err := gDBAsset.Exec(sql, uid, 0, 0)
	return err
}


func CheckTXID(txid int64)bool{
	row,err := gDBAsset.QueryRow("select count(1) as c from recent_tx_ids where txid = ?",txid)
	if err != nil {
		logger.Error("query row error ",err.Error())
		return false
	}
	return utils.Str2Int(row["c"])>0
}