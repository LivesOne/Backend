package common

import (
	"database/sql"
	"errors"
	_ "fmt"
	_ "github.com/go-sql-driver/mysql"
	"servlets/constants"
	"utils"
	"utils/config"
	_ "utils/config"
	"utils/db_factory"
	"utils/logger"
)

const (
	CONV_LVT = 10000 * 10000
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
		logger.Debug("connection ", db_config_asset.DBHost, db_config_asset.DBDatabase, "database successful")
	} else {
		logger.Error(gDBAsset.Err())
		return gDBAsset.Err()
	}

	return nil
}

func QueryReward(uid int64) (*Reward, error) {
	if uid == 0 {
		return nil, errors.New("uid is zero")
	}
	row, err := gDBAsset.QueryRow("select total,lastday,lastmodify from user_reward where uid = ?", uid)
	if err != nil {
		logger.Error("query db error ", err.Error())
		return nil, err
	}
	resReward := &Reward{
		Uid: uid,
	}
	if row != nil {
		resReward.Total = utils.Str2Int64(row["total"])
		resReward.Yesterday = utils.Str2Int64(row["lastday"])
		resReward.Lastmodify = utils.Str2Int64(row["lastmodify"])

	}
	return resReward, err

}

func QueryBalance(uid int64) (int64, error) {
	row, err := gDBAsset.QueryRow("select balance from user_asset where uid = ?", uid)
	if err != nil {
		logger.Error("query db error ", err.Error())
	}

	if row != nil {
		return utils.Str2Int64(row["balance"]), nil
	}
	return 0, err
}

func TransAccountLvt(txid, from, to, value int64) (bool, int) {
	//检测资产初始化情况
	//from 的资产如果没有初始化，初始化并返回false--》 上层检测到false会返回余额不足
	f, c := CheckAndInitAsset(from)
	if !f {
		return f, c
	}
	ts := utils.GetTimestamp13()

	tx, err := gDBAsset.Begin()
	if err != nil {
		logger.Error("db pool begin error ", err.Error())
		return false, constants.TRANS_ERR_SYS
	}
	tx.Exec("select * from user_asset where uid in (?,?) for update", from, to)

	//资产冻结状态校验，如果status是0 返回true 继续执行，status ！= 0 账户冻结，返回错误
	if !CheckAssetLimeted(from, tx) {
		tx.Rollback()
		return false, constants.TRANS_ERR_ASSET_LIMITED
	}

	//查询转出账户余额是否满足需要
	//var balance int64
	//row := tx.QueryRow("select balance from user_asset where uid  = ?", from)
	//row.Scan(&balance)
	//
	//if balance < value {
	//	tx.Rollback()
	//	return false, constants.TRANS_ERR_INSUFFICIENT_BALANCE
	//}

	//查询转出账户余额是否满足需要 使用新的校验方法，考虑到锁仓的问题
	if !ckeckBalance(from,value,tx) {
		tx.Rollback()
		return false, constants.TRANS_ERR_INSUFFICIENT_BALANCE
	}


	//扣除转出方balance
	info1, err1 := tx.Exec("update user_asset set balance = balance - ?,lastmodify = ? where uid = ?", value, ts, from)
	if err1 != nil {
		logger.Error("sql error ", err1.Error())
		tx.Rollback()
		return false, constants.TRANS_ERR_SYS
	}
	//update 以后校验修改记录条数，如果为0 说明初始化部分出现问题，返回错误
	rsa, _ := info1.RowsAffected()
	if rsa == 0 {
		logger.Error("update user balance error RowsAffected ", rsa, " can not find user  ", from, "")
		tx.Rollback()
		return false, constants.TRANS_ERR_SYS
	}
	//增加目标的balance
	info2, err2 := tx.Exec("update user_asset set balance = balance + ?,lastmodify = ? where uid = ?", value, ts, to)
	if err2 != nil {
		logger.Error("sql error ", err2.Error())
		tx.Rollback()
		return false, constants.TRANS_ERR_SYS
	}
	//update 以后校验修改记录条数，如果为0 说明初始化部分出现问题，返回错误
	rsa, _ = info2.RowsAffected()
	if rsa == 0 {
		logger.Error("update user balance error RowsAffected ", rsa, " can not find user  ", to, "")
		tx.Rollback()
		return false, constants.TRANS_ERR_SYS
	}
	//txid 写入数据库
	_, e := InsertTXID(txid, tx)

	if e != nil {
		logger.Error("sql error ", e.Error())
		tx.Rollback()
		return false, constants.TRANS_ERR_SYS
	}

	tx.Commit()

	return true, constants.TRANS_ERR_SUCC
}

func CheckAndInitAsset(uid int64) (bool, int) {
	//初始化资产
	res, err := InsertAsset(uid)
	if err != nil {
		logger.Error("init asset error ", err.Error())
		return false, constants.TRANS_ERR_SYS
	}

	rowsCount, err := res.RowsAffected()
	if err != nil {
		logger.Error("get RowsAffected error ", err.Error())
		return false, constants.TRANS_ERR_SYS
	}

	_, err = InsertReward(uid)
	if err != nil {
		logger.Error("init reward error ", err.Error())
		return false, constants.TRANS_ERR_SYS
	}

	if rowsCount == 0 {
		return true, constants.TRANS_ERR_SUCC
	}
	return false, constants.TRANS_ERR_INSUFFICIENT_BALANCE
}

func InsertTXID(txid int64, tx *sql.Tx) (sql.Result, error) {
	res, err := tx.Exec("Insert into recent_tx_ids values (?)", txid)
	if err != nil {
		logger.Error("query error ", err.Error())
	}
	return res, err

}

func RemoveTXID(txid int64) error {
	_, err := gDBAsset.Exec("delete from recent_tx_ids where txid = ?", txid)
	if err != nil {
		logger.Error("query error ", err.Error())
	}
	return err
}

func InsertReward(uid int64) (sql.Result, error) {
	sql := "insert ignore into user_reward (uid,total,lastday,lastmodify) values (?,?,?,?) "
	return gDBAsset.Exec(sql, uid, 0, 0, 0)
}

func InsertAsset(uid int64) (sql.Result, error) {
	sql := "insert ignore into user_asset (uid,balance,lastmodify) values (?,?,?) "
	return gDBAsset.Exec(sql, uid, 0, 0)
}

func CheckTXID(txid int64) bool {
	row, err := gDBAsset.QueryRow("select count(1) as c from recent_tx_ids where txid = ?", txid)
	if err != nil {
		logger.Error("query row error ", err.Error())
		return false
	}
	return utils.Str2Int(row["c"]) > 0
}

func CheckTansTypeFromUid(uid int64, transType int) bool {
	row, err := gDBAsset.QueryRow("select uid from livesone_account where id = ?", transType)
	if err != nil {
		logger.Error("query row error ", err.Error())
		return false
	}
	if row == nil {
		return false
	}
	return utils.Str2Int64(row["uid"]) == uid
}

func CheckAssetLimeted(uid int64, tx *sql.Tx) bool {
	row := tx.QueryRow("select status from user_asset where uid = ?", uid)
	status := -1
	err := row.Scan(&status)
	if err != nil {
		logger.Error("query row error ", err.Error())
		return false
	}
	return status == constants.ASSET_STATUS_DEF
}

func GetUserAssetTranslevelByUid(uid int64) int {
	res, err := gDBAsset.QueryRow("select trader_level from user_restrict where uid = ?", uid)
	if err != nil {
		logger.Error("cannot get trans level ", err.Error())
		return 0
	}
	if res == nil {
		logger.Info("can not find trans level by uid ", uid)
		return 0
	}
	return utils.Str2Int(res["trader_level"])
}

func ckeckBalance(uid int64,value int64, tx *sql.Tx)bool{
	var balance int64
	var locked int64
	row := tx.QueryRow("select balance,locked from user_asset where uid  = ?", uid)
	row.Scan(&balance,&locked)
	return balance > 0 && (balance-locked) > value
}

func CreateAssetLock(assetLock *AssetLock)(bool,int){

	tx, err := gDBAsset.Begin()
	if err != nil {
		logger.Error("db pool begin error ", err.Error())
		return false, constants.TRANS_ERR_SYS
	}
	//锁定记录
	tx.Exec("select * from user_asset where uid = ? for update", assetLock.Uid)

	//查询转出账户余额是否满足需要
	if !ckeckBalance(assetLock.Uid,assetLock.ValueInt,tx) {
		tx.Rollback()
		return false, constants.TRANS_ERR_INSUFFICIENT_BALANCE
	}


	//资产冻结状态校验，如果status是0 返回true 继续执行，status ！= 0 账户冻结，返回错误
	if !CheckAssetLimeted(assetLock.Uid, tx) {
		tx.Rollback()
		return false, constants.TRANS_ERR_ASSET_LIMITED
	}

	//修改资产数据
	//锁仓算力大于500时 给500
	updSql := `update
					user_asset
			   set
			   		locked = locked + ?,
			   		lock_hr = (case when (lock_hr + ? > ? ) then ? else (lock_hr + ?) end ),
			   		lastmodify = ?
			   where
			   		uid = ?`
	updParams := []interface{}{
		assetLock.ValueInt,
		assetLock.Hashrate,
		constants.ASSET_LOCK_MAX_VALUE,
		constants.ASSET_LOCK_MAX_VALUE,
		assetLock.Hashrate,
		assetLock.Begin,
		assetLock.Uid,
	}
	info1, err1 := tx.Exec(updSql, updParams...)
	if err1 != nil {
		logger.Error("sql error ", err1.Error())
		tx.Rollback()
		return false, constants.TRANS_ERR_SYS
	}
	//update 以后校验修改记录条数，如果为0 说明初始化部分出现问题，返回错误
	rsa, _ := info1.RowsAffected()
	if rsa == 0 {
		logger.Error("update user balance error RowsAffected ", rsa, " can not find user  ", assetLock.Uid, "")
		tx.Rollback()
		return false, constants.TRANS_ERR_SYS
	}

	sql := "insert into user_asset_lock (uid,value,month,hashrate,begin,end) values (?,?,?,?,?,?)"
	params := []interface{}{
		assetLock.Uid,
		assetLock.ValueInt,
		assetLock.Month,
		assetLock.Hashrate,
		assetLock.Begin,
		assetLock.End,
	}
	res,err := tx.Exec(sql,params...)
	if err != nil {
		logger.Error("create asset lock error",err.Error())
		tx.Rollback()
		return false,constants.TRANS_ERR_SYS
	}
	id,err := res.LastInsertId()
	if err != nil {
		logger.Error("get last insert id error",err.Error())
		tx.Rollback()
		return false,constants.TRANS_ERR_SYS
	}
	tx.Commit()
	assetLock.Id = id
	assetLock.IdStr = utils.Int642Str(id)
	return true, constants.TRANS_ERR_SUCC
}


func QueryAssetLockList(uid int64)[]*AssetLock{
	res := gDBAsset.Query("select * from user_asset_lock where uid = ? and end > ?",uid,utils.GetTimestamp13())
	if res == nil {
		return nil
	}
	return convAssetLockList(res)
}

func QueryAssetLock(id int64)*AssetLock{
	res,err := gDBAsset.QueryRow("select * from user_asset_lock where id = ?",id)
	if err != nil {
		logger.Error("query asset lock error",err.Error())
		return nil
	}
	return convAssetLock(res)
}

func convAssetLock(al map[string]string)*AssetLock{
	if al == nil {
		return nil
	}
	alres := AssetLock{
		Id:       utils.Str2Int64(al["id"]),
		Uid:      utils.Str2Int64(al["uid"]),
		ValueInt: utils.Str2Int64(al["value"]),
		Month:    utils.Str2Int(al["month"]),
		Hashrate: utils.Str2Int(al["hashrate"]),
		Begin:    utils.Str2Int64(al["begin"]),
		End:      utils.Str2Int64(al["end"]),
	}
	alres.Value = utils.LVTintToFloatStr(alres.ValueInt)
	alres.IdStr = utils.Int642Str(alres.Id)
	return &alres
}

func convAssetLockList(list []map[string]string)[]*AssetLock{
	listRes := make([]*AssetLock,0)
	for _,v := range list {
		listRes = append(listRes,convAssetLock(v))
	}
	return listRes
}


func execRemoveAssetLock(txid int64,assetLock *AssetLock,penaltyMoney int64,tx *sql.Tx)(bool,int){
	//锁定记录
	ts := utils.TXIDToTimeStamp13(txid)
	to := config.GetConfig().PenaltyMoneyAccountUid
	tx.Exec("select * from user_asset where uid in (?,?) for update", assetLock.Uid,to)

	//资产冻结状态校验，如果status是0 返回true 继续执行，status ！= 0 账户冻结，返回错误
	if !CheckAssetLimeted(assetLock.Uid, tx) {
		return false, constants.TRANS_ERR_ASSET_LIMITED
	}

	//修改资产数据
	//锁仓算力大于500时 给500
	updSql := `update
					user_asset
			   set
			   		balance = balance - ?,
			   		locked = locked - ?,
			   		lastmodify = ?
			   where
			   		uid = ?`
	updParams := []interface{}{
		penaltyMoney,
		assetLock.ValueInt,
		ts,
		assetLock.Uid,
	}
	info1, err := tx.Exec(updSql, updParams...)
	if err != nil {
		logger.Error("sql error ", err.Error())
		return false, constants.TRANS_ERR_SYS
	}
	//update 以后校验修改记录条数，如果为0 说明初始化部分出现问题，返回错误
	rsa, _ := info1.RowsAffected()
	if rsa == 0 {
		logger.Error("update user balance error RowsAffected ", rsa, " can not find user  ", assetLock.Uid, "")
		return false, constants.TRANS_ERR_SYS
	}

	//增加目标的balance to为系统配置账户
	info2, err2 := tx.Exec("update user_asset set balance = balance + ?,lastmodify = ? where uid = ?", assetLock.ValueInt, ts, to)
	if err2 != nil {
		logger.Error("sql error ", err2.Error())
		return false, constants.TRANS_ERR_SYS
	}
	//update 以后校验修改记录条数，如果为0 说明初始化部分出现问题，返回错误
	rsa, _ = info2.RowsAffected()
	if rsa == 0 {
		logger.Error("update user balance error RowsAffected ", rsa, " can not find user  ", to, "")
		return false, constants.TRANS_ERR_SYS
	}


	_,err = tx.Exec("delete from user_asset_lock where id = ?",assetLock.Id)
	if err != nil {
		logger.Error("create asset lock error",err.Error())
		return false,constants.TRANS_ERR_SYS
	}


	//修改资产解冻之后，要重新计算应该有的hashrate
	var hr int
	row := tx.QueryRow("select sum(hashreat) from user_asset_lock where uid = ?",assetLock.Uid)

	err = row.Scan(&hr)
	if err != nil {
		logger.Error("query asset lock hashrate error",err.Error())
		return false,constants.TRANS_ERR_SYS
	}
	if hr > constants.ASSET_LOCK_MAX_VALUE {
		hr = constants.ASSET_LOCK_MAX_VALUE
	}

	_,err = tx.Exec("update user_asset set lock_hr = ? where uid = ?",hr,assetLock.Uid)
	if err != nil {
		logger.Error("sql error ", err.Error())
		return false, constants.TRANS_ERR_SYS
	}
	return true,constants.TRANS_ERR_SUCC
}

func RemoveAssetLock(txid int64,assetLock *AssetLock,penaltyMoney int64)(bool,int){
	ts := utils.TXIDToTimeStamp13(txid)
	tx, err := gDBAsset.Begin()
	if err != nil {
		logger.Error("db pool begin error ", err.Error())
		return false, constants.TRANS_ERR_SYS
	}


	if ok,e := execRemoveAssetLock(txid,assetLock,penaltyMoney,tx);!ok{
		tx.Rollback()
		return false,e
	}

	//加入交易记录，不成功的话回滚并返回系统错误
	txh := &DTTXHistory{
		Id:     txid,
		Status: constants.TX_STATUS_DEFAULT,
		Type:   constants.TX_TYPE_PENALTY_MONEY,
		From:   assetLock.Uid,
		To:     config.GetConfig().PenaltyMoneyAccountUid,
		Value:  assetLock.ValueInt,
		Ts:     ts,
		Code:   constants.TX_CODE_SUCC,
		Remark: assetLock,
	}
	err = InsertCommited(txh)
	if err != nil {
		logger.Error("insert mongo  error ", err.Error())
		tx.Rollback()
		return false, constants.TRANS_ERR_SYS
	}

	err = tx.Commit()
	if err != nil {
		logger.Error("mysql commit  error ", err.Error())
		return false, constants.TRANS_ERR_SYS
	}

	return true, constants.TRANS_ERR_SUCC
}