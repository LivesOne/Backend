package common

import (
	"database/sql"
	sqlBase "database/sql"
	"errors"
	"fmt"
	"github.com/garyburd/redigo/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/shopspring/decimal"
	"gitlab.maxthon.net/cloud/livesone-user-micro/src/proto"
	"servlets/constants"
	"servlets/rpc"
	"strings"
	"time"
	"utils"
	"utils/config"
	"utils/db_factory"
	"utils/logger"
)

const (
	CONV_LVT          = 10000 * 10000
	DAY_QUOTA_TYPE    = 1
	CASUAL_QUOTA_TYPE = 2
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
	row, err := gDBAsset.QueryRow("select total,lastday,lastmodify,days from user_reward where uid = ?", uid)
	if err != nil {
		logger.Error("query db:user_reward error ", err.Error())
		return nil, err
	}
	resReward := &Reward{
		Uid: uid,
	}
	if row != nil {
		resReward.Total = utils.Str2Int64(row["total"])
		resReward.Yesterday = utils.Str2Int64(row["lastday"])
		resReward.Lastmodify = utils.Str2Int64(row["lastmodify"])
		resReward.Days = utils.Str2Int(row["days"])

	}
	return resReward, err

}

func QueryLvtcReward(uid int64) (*Reward, error) {
	if uid == 0 {
		return nil, errors.New("uid is zero")
	}
	resReward := &Reward{
		Uid: uid,
	}
	row, err := gDBAsset.QueryRow("select total,lastday,lastmodify,days from user_reward_lvtc where uid = ?", uid)
	if err != nil {
		logger.Error("query db:user_reward_lvtc error ", err.Error())
		return nil, err
	}
	if row != nil {
		resReward.Total = utils.Str2Int64(row["total"])
		resReward.Yesterday = utils.Str2Int64(row["lastday"])
		resReward.Lastmodify = utils.Str2Int64(row["lastmodify"])
		resReward.Days = utils.Str2Int(row["days"])

	} else {
		// no record
		// query table: user_reward
		reward, err := QueryReward(uid)
		if err != nil {
			return nil, err
		}
		if reward != nil && reward.Lastmodify != 0 {
			resReward.Days = reward.Days
			resReward.Lastmodify = reward.Lastmodify
		}
	}
	return resReward, err

}

func QueryBalance(uid int64) (int64, int64, int64, int64, int, error) {
	row, err := gDBAsset.QueryRow("select balance,locked,income,lastmodify,status from user_asset where uid = ?", uid)
	if err != nil {
		logger.Error("query db error ", err.Error())
	}

	if row != nil {
		return utils.Str2Int64(row["balance"]), utils.Str2Int64(row["locked"]), utils.Str2Int64(row["income"]), utils.Str2Int64(row["lastmodify"]), utils.Str2Int(row["status"]), nil
	}
	return 0, 0, 0, 0, 0, err
}

func QueryBalanceLvtc(uid int64) (int64, int64, int64, int64, int, error) {
	row, err := gDBAsset.QueryRow("select balance,locked,income,lastmodify,status from user_asset_lvtc where uid = ?", uid)
	if err != nil {
		logger.Error("query db error ", err.Error())
	}

	if row != nil {
		return utils.Str2Int64(row["balance"]), utils.Str2Int64(row["locked"]), utils.Str2Int64(row["income"]), utils.Str2Int64(row["lastmodify"]), utils.Str2Int(row["status"]), nil
	}
	return 0, 0, 0, 0, 0, err
}

func QueryBalanceEth(uid int64) (int64, int64, int64, int64, int, error) {
	row, err := gDBAsset.QueryRow("select balance,locked,income,lastmodify,status from user_asset_eth where uid = ?", uid)
	if err != nil {
		logger.Error("query db error ", err.Error())
	}

	if row != nil {
		return utils.Str2Int64(row["balance"]), utils.Str2Int64(row["locked"]), utils.Str2Int64(row["income"]), utils.Str2Int64(row["lastmodify"]), utils.Str2Int(row["status"]), nil
	}
	return 0, 0, 0, 0, 0, err
}

func QueryBalanceEos(uid int64) (int64, int64, int64, int64, int, error) {
	row, err := gDBAsset.QueryRow("select balance,locked,income,lastmodify,status from user_asset_eos where uid = ?", uid)
	if err != nil {
		logger.Error("query db error ", err.Error())
	}

	if row != nil {
		return utils.Str2Int64(row["balance"]), utils.Str2Int64(row["locked"]), utils.Str2Int64(row["income"]), utils.Str2Int64(row["lastmodify"]), utils.Str2Int(row["status"]), nil
	}
	return 0, 0, 0, 0, 0, err
}

func QueryBalanceBtc(uid int64) (int64, int64, int64, int64, int, error) {
	row, err := gDBAsset.QueryRow("select balance,locked,income,lastmodify,status from user_asset_btc where uid = ?", uid)
	if err != nil {
		logger.Error("query db error ", err.Error())
	}

	if row != nil {
		return utils.Str2Int64(row["balance"]), utils.Str2Int64(row["locked"]), utils.Str2Int64(row["income"]), utils.Str2Int64(row["lastmodify"]), utils.Str2Int(row["status"]), nil
	}
	return 0, 0, 0, 0, 0, err
}

func QueryBalanceBsv(uid int64) (int64, int64, int64, int64, int, error) {
	row, err := gDBAsset.QueryRow("select balance,locked,income,lastmodify,status from user_asset_bsv where uid = ?", uid)
	if err != nil {
		logger.Error("query db error ", err.Error())
	}

	if row != nil {
		return utils.Str2Int64(row["balance"]), utils.Str2Int64(row["locked"]), utils.Str2Int64(row["income"]), utils.Str2Int64(row["lastmodify"]), utils.Str2Int(row["status"]), nil
	}
	return 0, 0, 0, 0, 0, err
}

func TransAccountLvt(tx *sql.Tx, dth *DTTXHistory) (bool, int) {
	f, c := TransAccountLvtByTx(dth.Id, dth.From, dth.To, dth.Value, tx)
	if !f {
		return f, c
	}
	err := InsertCommited(dth)
	if CheckDup(err) {
		return true, constants.TRANS_ERR_SUCC
	}
	return false, constants.TRANS_ERR_SYS
}

func TransAccountLvtc(tx *sql.Tx, dth *DTTXHistory) (bool, int) {
	f, c := TransAccountLvtcByTx(dth.Id, dth.From, dth.To, dth.Value, tx)
	if !f {
		return f, c
	}
	err := InsertLVTCCommited(dth)
	if CheckDup(err) {
		return true, constants.TRANS_ERR_SUCC
	}
	return false, constants.TRANS_ERR_SYS
}

func TransAccountLvtByTx(txid, from, to, value int64, tx *sql.Tx) (bool, int) {
	tx.Exec("select * from user_asset where uid in (?,?) for update", from, to)

	ts := utils.GetTimestamp13()

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
	if !ckeckBalance(from, value, tx) {
		return false, constants.TRANS_ERR_INSUFFICIENT_BALANCE
	}
	//资产冻结状态校验，如果status是0 返回true 继续执行，status ！= 0 账户冻结，返回错误
	if !CheckAssetLimeted(from, tx) {
		return false, constants.TRANS_ERR_ASSET_LIMITED
	}

	//扣除转出方balance
	info1, err1 := tx.Exec("update user_asset set balance = balance - ?,lastmodify = ? where uid = ?", value, ts, from)
	if err1 != nil {
		logger.Error("sql error ", err1.Error())
		return false, constants.TRANS_ERR_SYS
	}
	//update 以后校验修改记录条数，如果为0 说明初始化部分出现问题，返回错误
	rsa, _ := info1.RowsAffected()
	if rsa == 0 {
		logger.Error("update user balance error RowsAffected ", rsa, " can not find user  ", from, "")
		return false, constants.TRANS_ERR_SYS
	}
	//增加目标的balance
	info2, err2 := tx.Exec("update user_asset set balance = balance + ?,lastmodify = ? where uid = ?", value, ts, to)
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
	//txid 写入数据库
	_, e := InsertTXID(txid, tx)

	if e != nil {
		logger.Error("sql error ", e.Error())
		return false, constants.TRANS_ERR_SYS
	}
	return true, constants.TRANS_ERR_SUCC
}

func TransAccountLvtcByTx(txid, from, to, value int64, tx *sql.Tx) (bool, int) {
	tx.Exec("select * from user_asset_lvtc where uid in (?,?) for update", from, to)

	ts := utils.GetTimestamp13()

	//查询转出账户余额是否满足需要 使用新的校验方法，考虑到锁仓的问题
	if !ckeckBalanceOfLvtc(from, value, tx) {
		return false, constants.TRANS_ERR_INSUFFICIENT_BALANCE
	}
	//资产冻结状态校验，如果status是0 返回true 继续执行，status ！= 0 账户冻结，返回错误
	if !CheckAssetLimetedOfLvtc(from, tx) {
		return false, constants.TRANS_ERR_ASSET_LIMITED
	}

	//扣除转出方balance
	info1, err1 := tx.Exec("update user_asset_lvtc set balance = balance - ?,lastmodify = ? where uid = ?", value, ts, from)
	if err1 != nil {
		logger.Error("sql error ", err1.Error())
		return false, constants.TRANS_ERR_SYS
	}
	//update 以后校验修改记录条数，如果为0 说明初始化部分出现问题，返回错误
	rsa, _ := info1.RowsAffected()
	if rsa == 0 {
		logger.Error("update user balance error RowsAffected ", rsa, " can not find user  ", from, "")
		return false, constants.TRANS_ERR_SYS
	}
	//增加目标的balance
	info2, err2 := tx.Exec("update user_asset_lvtc set balance = balance + ?,lastmodify = ? where uid = ?", value, ts, to)
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
	//txid 写入数据库
	_, e := InsertTXID(txid, tx)
	if e != nil {
		logger.Error("sql error ", e.Error())
		return false, constants.TRANS_ERR_SYS
	}
	return true, constants.TRANS_ERR_SUCC
}

func ConvAccountLvtcBsvByTx(txid, systemUid, to, lvtc, bsv int64, tx *sql.Tx) (bool, int) {
	tx.Exec("select * from user_asset_lvtc where uid in (?,?) for update", systemUid, to)

	ts := utils.GetTimestamp13()
	//扣除转出方balance
	info0, err0 := tx.Exec("update user_asset_lvtc set balance = balance - ?," +
		"lastmodify = ? where uid = ?", lvtc, ts, to)
	if err0 != nil {
		logger.Error("sql error ", err0.Error())
		return false, constants.TRANS_ERR_SYS
	}
	//update 以后校验修改记录条数，如果为0 说明初始化部分出现问题，返回错误
	rsa, _ := info0.RowsAffected()
	if rsa == 0 {
		logger.Error("update user lvtc balance error RowsAffected ", rsa, " can not find user  ",
			to, "")
		return false, constants.TRANS_ERR_SYS
	}

	//扣除转出方balance
	info1, err1 := tx.Exec("update user_asset_lvtc set balance = balance + ?,lastmodify = ? where uid = ?", lvtc, ts, systemUid)
	if err1 != nil {
		logger.Error("sql error ", err1.Error())
		return false, constants.TRANS_ERR_SYS
	}
	//update 以后校验修改记录条数，如果为0 说明初始化部分出现问题，返回错误
	rsa, _ = info1.RowsAffected()
	if rsa == 0 {
		logger.Error("update user lvtc balance error RowsAffected ", rsa, " can not find user  ",
			systemUid, "")
		return false, constants.TRANS_ERR_SYS
	}
	//增加目标的balance
	info2, err2 := tx.Exec("update user_asset_bsv set balance = balance + ?," +
		"lastmodify = ? where uid = ?", bsv, ts, to)
	if err2 != nil {
		logger.Error("sql error ", err2.Error())
		return false, constants.TRANS_ERR_SYS
	}
	//update 以后校验修改记录条数，如果为0 说明初始化部分出现问题，返回错误
	rsa, _ = info2.RowsAffected()
	if rsa == 0 {
		logger.Error("update user bsv balance error RowsAffected ", rsa, " can not find user  ", to,
			"")
		return false, constants.TRANS_ERR_SYS
	}
	//txid 写入数据库
	_, e := InsertTXID(txid, tx)
	if e != nil {
		logger.Error("sql error ", e.Error())
		return false, constants.TRANS_ERR_SYS
	}
	return true, constants.TRANS_ERR_SUCC
}

func ConvAccountLvtcByTx(txid, systemUid, to, lvt, lvtc int64, tx *sql.Tx) (bool, int) {
	tx.Exec("select * from user_asset_lvtc where uid in (?,?) for update", systemUid, to)

	ts := utils.GetTimestamp13()
	sysValue := lvt - lvtc
	//扣除转出方balance
	info1, err1 := tx.Exec("update user_asset_lvtc set balance = balance + ?,lastmodify = ? where uid = ?", sysValue, ts, systemUid)
	if err1 != nil {
		logger.Error("sql error ", err1.Error())
		return false, constants.TRANS_ERR_SYS
	}
	//update 以后校验修改记录条数，如果为0 说明初始化部分出现问题，返回错误
	rsa, _ := info1.RowsAffected()
	if rsa == 0 {
		logger.Error("update user balance error RowsAffected ", rsa, " can not find user  ", systemUid, "")
		return false, constants.TRANS_ERR_SYS
	}
	//增加目标的balance
	info2, err2 := tx.Exec("update user_asset_lvtc set balance = balance + ?,lastmodify = ? where uid = ?", lvtc, ts, to)
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
	//txid 写入数据库
	_, e := InsertTXID(txid, tx)
	if e != nil {
		logger.Error("sql error ", e.Error())
		return false, constants.TRANS_ERR_SYS
	}
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
	_, err = InsertAssetLvtc(uid)
	if err != nil {
		logger.Error("init lvtc_asset error ", err.Error())
		return false, constants.TRANS_ERR_SYS
	}

	_, err = InsertAssetEth(uid)
	if err != nil {
		logger.Error("init asset eth error ", err.Error())
		return false, constants.TRANS_ERR_SYS
	}

	_, err = InsertAssetEos(uid)
	if err != nil {
		logger.Error("init asset eos error ", err.Error())
		return false, constants.TRANS_ERR_SYS
	}

	_, err = InsertAssetBtc(uid)
	if err != nil {
		logger.Error("init asset btc error ", err.Error())
		return false, constants.TRANS_ERR_SYS
	}

	_, err = InsertAssetBsv(uid)
	if err != nil {
		logger.Error("init bsv eth error ", err.Error())
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

func InsertRewardLvtc(uid int64) (sql.Result, error) {
	sql := "insert ignore into user_reward_lvtc (uid,total,lastday,lastmodify) values (?,?,?,?) "
	return gDBAsset.Exec(sql, uid, 0, 0, 0)
}

func InsertAsset(uid int64) (sql.Result, error) {
	sql := "insert ignore into user_asset (uid,balance,lastmodify) values (?,?,?) "
	return gDBAsset.Exec(sql, uid, 0, 0)
}

func InsertAssetLvtc(uid int64) (sql.Result, error) {
	sql := "insert ignore into user_asset_lvtc (uid,balance,lastmodify) values (?,?,?) "
	return gDBAsset.Exec(sql, uid, 0, 0)
}
func InsertAssetEth(uid int64) (sql.Result, error) {
	sql := "insert ignore into user_asset_eth (uid,balance,lastmodify) values (?,?,?) "
	return gDBAsset.Exec(sql, uid, 0, 0)
}

func InsertAssetEos(uid int64) (sql.Result, error) {
	sql := "insert ignore into user_asset_eos (uid,balance,lastmodify) values (?,?,?) "
	return gDBAsset.Exec(sql, uid, 0, 0)
}

func InsertAssetBtc(uid int64) (sql.Result, error) {
	sql := "insert ignore into user_asset_btc (uid,balance,lastmodify) values (?,?,?) "
	return gDBAsset.Exec(sql, uid, 0, 0)
}

func InsertAssetBsv(uid int64) (sql.Result, error) {
	sql := "insert ignore into user_asset_bsv (uid,balance,lastmodify) values (?,?,?) "
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

func checkAssetLimeted(tb string, uid int64, tx *sql.Tx) bool {
	row := tx.QueryRow(fmt.Sprintf("select status from %s where uid = ?", tb), uid)
	status := -1
	err := row.Scan(&status)
	if err != nil {
		logger.Error("query row error ", err.Error())
		return false
	}
	return status == constants.ASSET_STATUS_DEF
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

func CheckAssetLimetedOfLvtc(uid int64, tx *sql.Tx) bool {
	row := tx.QueryRow("select status from user_asset_lvtc where uid = ?", uid)
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

func ckeckBalance(uid int64, value int64, tx *sql.Tx) bool {
	var balance, income int64
	var locked int64
	row := tx.QueryRow("select balance,locked,income from user_asset where uid  = ?", uid)
	row.Scan(&balance, &locked, &income)
	logger.Info("balance", balance, "locked", locked, "income", income)
	return balance > 0 && (balance-locked-income) >= value
}

func ckeckBalanceOfLvtc(uid int64, value int64, tx *sql.Tx) bool {
	var balance, income int64
	var locked int64
	row := tx.QueryRow("select balance,locked,income from user_asset_lvtc where uid  = ?", uid)
	row.Scan(&balance, &locked, &income)
	logger.Info("balance", balance, "locked", locked, "income", income)
	return balance > 0 && (balance-locked-income) >= value
}

func ckeckEthBalance(uid int64, value int64, tx *sql.Tx) bool {
	var balance, income int64
	var locked int64
	row := tx.QueryRow("select balance,locked,income from user_asset_eth where uid  = ?", uid)
	row.Scan(&balance, &locked, &income)
	logger.Info("balance", balance, "locked", locked, "income", income)
	return balance > 0 && (balance-locked-income) >= value
}

func ckeckEosBalance(uid int64, value int64, tx *sql.Tx) bool {
	var balance, income int64
	var locked int64
	row := tx.QueryRow("select balance,locked,income from user_asset_eos where uid  = ?", uid)
	row.Scan(&balance, &locked, &income)
	logger.Info("balance", balance, "locked", locked, "income", income)
	return balance > 0 && (balance-locked-income) >= value
}

func ckeckBtcBalance(uid int64, value int64, tx *sql.Tx) bool {
	var balance, income int64
	var locked int64
	row := tx.QueryRow("select balance,locked,income from user_asset_btc where uid  = ?", uid)
	row.Scan(&balance, &locked, &income)
	logger.Info("balance", balance, "locked", locked, "income", income)
	return balance > 0 && (balance-locked-income) >= value
}

func CreateAssetLockByTx(assetLock *AssetLockLvtc, tx *sql.Tx) (bool, int) {
	//锁定记录
	tx.Exec("select * from user_asset_lvtc where uid = ? for update", assetLock.Uid)

	//查询转出账户余额是否满足需要
	if !ckeckBalanceOfLvtc(assetLock.Uid, assetLock.ValueInt, tx) {
		return false, constants.TRANS_ERR_INSUFFICIENT_BALANCE
	}

	//资产冻结状态校验，如果status是0 返回true 继续执行，status ！= 0 账户冻结，返回错误
	if !CheckAssetLimetedOfLvtc(assetLock.Uid, tx) {
		return false, constants.TRANS_ERR_ASSET_LIMITED
	}

	//修改资产数据
	//锁仓算力大于500时 给500
	updSql := `update
					user_asset_lvtc
			   set
			   		locked = locked + ?,
			   		lastmodify = ?
			   where
			   		uid = ?`
	updParams := []interface{}{
		assetLock.ValueInt,
		assetLock.Begin,
		assetLock.Uid,
	}
	info1, err1 := tx.Exec(updSql, updParams...)
	if err1 != nil {
		logger.Error("sql error ", err1.Error())
		return false, constants.TRANS_ERR_SYS
	}
	//update 以后校验修改记录条数，如果为0 说明初始化部分出现问题，返回错误
	rsa, _ := info1.RowsAffected()
	if rsa == 0 {
		logger.Error("update user balance error RowsAffected ", rsa, " can not find user  ", assetLock.Uid, "")
		return false, constants.TRANS_ERR_SYS
	}

	sql := "insert into user_asset_lock (uid,value,month,hashrate,begin,end,currency,allow_unlock) values (?,?,?,?,?,?,?,?)"
	params := []interface{}{
		assetLock.Uid,
		assetLock.ValueInt,
		assetLock.Month,
		assetLock.Hashrate,
		assetLock.Begin,
		assetLock.End,
		assetLock.Currency,
		assetLock.AllowUnlock,
	}
	res, err := tx.Exec(sql, params...)
	if err != nil {
		logger.Error("create asset lock error", err.Error())
		return false, constants.TRANS_ERR_SYS
	}
	id, err := res.LastInsertId()
	if err != nil {
		logger.Error("get last insert id error", err.Error())
		return false, constants.TRANS_ERR_SYS
	}

	assetLock.Id = id
	assetLock.IdStr = utils.Int642Str(id)

	if ok, code := updLockAssetHashRate(assetLock.Uid, tx); !ok {
		return ok, code
	}
	return true, constants.TRANS_ERR_SUCC
}

func CreateAssetLock(assetLock *AssetLockLvtc) (bool, int) {

	tx, err := gDBAsset.Begin()
	if err != nil {
		logger.Error("db pool begin error ", err.Error())
		return false, constants.TRANS_ERR_SYS
	}

	if ok, e := CreateAssetLockByTx(assetLock, tx); !ok {
		tx.Rollback()
		return ok, e
	}

	tx.Commit()
	return true, constants.TRANS_ERR_SUCC
}

func UpgradeAssetLock(assetLock *AssetLock) (bool, int) {

	tx, err := gDBAsset.Begin()
	if err != nil {
		logger.Error("db pool begin error ", err.Error())
		return false, constants.TRANS_ERR_SYS
	}
	//锁定记录
	tx.Exec("select * from user_asset where uid = ? for update", assetLock.Uid)

	//查询转出账户余额是否满足需要
	if !ckeckBalance(assetLock.Uid, assetLock.ValueInt, tx) {
		tx.Rollback()
		return false, constants.TRANS_ERR_INSUFFICIENT_BALANCE
	}

	//资产冻结状态校验，如果status是0 返回true 继续执行，status ！= 0 账户冻结，返回错误
	if !CheckAssetLimeted(assetLock.Uid, tx) {
		tx.Rollback()
		return false, constants.TRANS_ERR_ASSET_LIMITED
	}

	sql := "insert into user_asset_lock (uid,value,month,hashrate,begin,end,type) values (?,?,?,?,?,?,?)"
	sql = `
		update user_asset_lock set
			month = ?,
			hashrate = ?,
			begin = ?,
			end = ?,
			type = ?
		where
		id = ? and uid = ?

	`

	params := []interface{}{
		assetLock.Month,
		assetLock.Hashrate,
		assetLock.Begin,
		assetLock.End,
		ASSET_LOCK_TYPE_DRAW,
		assetLock.Id,
		assetLock.Uid,
	}
	_, err = tx.Exec(sql, params...)
	if err != nil {
		logger.Error("create asset lock error", err.Error())
		tx.Rollback()
		return false, constants.TRANS_ERR_SYS
	}

	if ok, code := updLockAssetHashRate(assetLock.Uid, tx); !ok {
		tx.Rollback()
		return ok, code
	}

	incomeCasual := int64(0)
	switch assetLock.Month {
	case 6:
		incomeCasual = assetLock.ValueInt / 2
	case 12:
		incomeCasual = assetLock.ValueInt
	default:
		tx.Rollback()
		logger.Error("month must by 6/12")
		return false, constants.TRANS_ERR_PARAM
	}

	if wr := InitUserWithdrawal(assetLock.Uid); wr != nil {
		if ok, _ := IncomeUserWithdrawalCasualQuota(assetLock.Uid, incomeCasual); !ok {
			tx.Rollback()
			return false, constants.TRANS_ERR_SYS
		}
	} else {
		tx.Rollback()
		return false, constants.TRANS_ERR_SYS
	}
	tx.Commit()
	return true, constants.TRANS_ERR_SUCC
}

func QueryAssetLockList(uid int64, currency string) []*AssetLockLvtc {
	res := gDBAsset.Query("select * from user_asset_lock where uid = ? and currency = ? order by id desc", uid, currency)
	if res == nil {
		return nil
	}
	return convAssetLockList(res)
}

func updLockAssetHashRate(uid int64, tx *sql.Tx) (bool, int) {
	//锁仓有所变动之后，要重新计算应该有的hashrate
	var hr int
	row := tx.QueryRow("select  if(sum(hashrate) is null,0,sum(hashrate)) as hr from user_asset_lock where uid = ?", uid)

	err := row.Scan(&hr)
	if err != nil {
		logger.Error("query asset lock hashrate error", err.Error())
		return false, constants.TRANS_ERR_SYS
	}
	if hr > constants.ASSET_LOCK_MAX_VALUE {
		hr = constants.ASSET_LOCK_MAX_VALUE
	}
	return updHashRate(uid, hr, USER_HASHRATE_TYPE_LOCK_ASSET, 0, 0, tx)
}

func updHashRate(uid int64, hr, hrType int, begin, end int64, tx *sql.Tx) (bool, int) {
	if tx == nil {
		var err error
		tx, err = gDBAsset.Begin()
		if err != nil {
			logger.Error("begin tx error ", err.Error())
			return false, constants.TRANS_ERR_SYS
		}
		defer tx.Commit()
	}

	hrc := tx.QueryRow("select id from user_hashrate where uid = ? and type = ? for update", uid, hrType)

	var hrId int64

	err := hrc.Scan(&hrId)
	if err != nil && err != sql.ErrNoRows {
		logger.Error("sql error ", err.Error())
		return false, constants.TRANS_ERR_SYS
	}

	if hrId > 0 {
		_, err := tx.Exec("update user_hashrate set hashrate = ?,begin = ? , end = ? where id = ?", hr, begin, end, hrId)
		if err != nil {
			logger.Error("sql error ", err.Error())
			return false, constants.TRANS_ERR_SYS
		}
	} else {
		_, err = tx.Exec("insert into user_hashrate(type,uid,hashrate,begin,end) values (?,?,?,?,?)", hrType, uid, hr, begin, end)
		if err != nil {
			logger.Error("sql error ", err.Error())
			return false, constants.TRANS_ERR_SYS
		}
	}

	return true, constants.TRANS_ERR_SUCC
}

func QueryAssetLock(id, uid int64) *AssetLock {
	res, err := gDBAsset.QueryRow("select * from user_asset_lock where id = ? and uid = ?", id, uid)
	if err != nil {
		logger.Error("query asset lock error", err.Error())
		return nil
	}
	return convAssetLock(res)
}

func QueryLvtcAssetLock(id, uid int64) *AssetLockLvtc {
	res, err := gDBAsset.QueryRow("select * from user_asset_lock where id = ? and uid = ?", id, uid)
	if err != nil {
		logger.Error("query asset lock error", err.Error())
		return nil
	}
	return convLvtcAssetLock(res)
}

func convAssetLock(al map[string]string) *AssetLock {
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
		Type:     utils.Str2Int(al["type"]),
	}
	alres.Value = utils.LVTintToFloatStr(alres.ValueInt)
	alres.IdStr = utils.Int642Str(alres.Id)
	return &alres
}

func convLvtcAssetLock(al map[string]string) *AssetLockLvtc {
	if al == nil {
		return nil
	}
	alres := AssetLockLvtc{
		Id:          utils.Str2Int64(al["id"]),
		Uid:         utils.Str2Int64(al["uid"]),
		ValueInt:    utils.Str2Int64(al["value"]),
		Month:       utils.Str2Int(al["month"]),
		Hashrate:    utils.Str2Int(al["hashrate"]),
		Begin:       utils.Str2Int64(al["begin"]),
		End:         utils.Str2Int64(al["end"]),
		Currency:    al["currency"],
		AllowUnlock: utils.Str2Int(al["allow_unlock"]),
		Income:      utils.Str2Int(al["income"]),
	}
	alres.Value = utils.LVTintToFloatStr(alres.ValueInt)
	alres.IdStr = utils.Int642Str(alres.Id)
	return &alres
}

func convAssetLockList(list []map[string]string) []*AssetLockLvtc {
	listRes := make([]*AssetLockLvtc, 0)
	for _, v := range list {
		listRes = append(listRes, convLvtcAssetLock(v))
	}
	return listRes
}

func execRemoveAssetLock(txid int64, assetLock *AssetLockLvtc, penaltyMoney int64, tx *sql.Tx) (bool, int) {
	//锁定记录
	ts := utils.TXIDToTimeStamp13(txid)
	to := config.GetConfig().PenaltyMoneyAccountUid
	tx.Exec("select * from user_asset_lvtc where uid in (?,?) for update", assetLock.Uid, to)

	//资产冻结状态校验，如果status是0 返回true 继续执行，status ！= 0 账户冻结，返回错误
	if !CheckAssetLimetedOfLvtc(assetLock.Uid, tx) {
		return false, constants.TRANS_ERR_ASSET_LIMITED
	}

	//修改资产数据
	//锁仓算力大于500时 给500
	var updSql string
	var updParams []interface{}
	if assetLock.Income == ASSET_INCOME_MINING {
		updSql = `update user_asset_lvtc
			   set balance = balance - ?,locked = locked - ?,income = income + ?,lastmodify = ?
			   where uid = ?`
		updParams = []interface{}{penaltyMoney, assetLock.ValueInt, assetLock.ValueInt - penaltyMoney, ts, assetLock.Uid}
	} else {
		updSql = `update user_asset_lvtc
			   set balance = balance - ?, locked = locked - ?, lastmodify = ?
			   where uid = ?`
		updParams = []interface{}{penaltyMoney, assetLock.ValueInt, ts, assetLock.Uid}
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
	info2, err2 := tx.Exec("update user_asset_lvtc set balance = balance + ?,lastmodify = ? where uid = ?", penaltyMoney, ts, to)
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

	_, err = tx.Exec("delete from user_asset_lock where id = ?", assetLock.Id)
	if err != nil {
		logger.Error("create asset lock error", err.Error())
		return false, constants.TRANS_ERR_SYS
	}

	if ok, code := updLockAssetHashRate(assetLock.Uid, tx); !ok {
		return ok, code
	}
	return true, constants.TRANS_ERR_SUCC
}

func RemoveAssetLock(txid int64, assetLock *AssetLockLvtc, penaltyMoney int64) (bool, int) {
	ts := utils.TXIDToTimeStamp13(txid)
	tx, err := gDBAsset.Begin()
	if err != nil {
		logger.Error("db pool begin error ", err.Error())
		return false, constants.TRANS_ERR_SYS
	}

	if ok, e := execRemoveAssetLock(txid, assetLock, penaltyMoney, tx); !ok {
		tx.Rollback()
		return false, e
	}

	//加入交易记录，不成功的话回滚并返回系统错误
	txh := &DTTXHistory{
		Id:     txid,
		Status: constants.TX_STATUS_DEFAULT,
		Type:   constants.TX_TYPE_PENALTY_MONEY,
		From:   assetLock.Uid,
		To:     config.GetConfig().PenaltyMoneyAccountUid,
		Value:  penaltyMoney,
		Ts:     ts,
		Code:   constants.TX_CODE_SUCC,
		Remark: assetLock,
	}
	err = InsertLVTCCommited(txh)
	if err != nil {
		logger.Error("insert mongo  error ", err.Error())
		tx.Rollback()
		return false, constants.TRANS_ERR_SYS
	}

	fromName, err := rpc.GetUserField(txh.From, microuser.UserField_NICKNAME)
	if err != nil {
		logger.Info("get uid:", txh.From, " nick name err,", err)
	}
	toName, err := rpc.GetUserField(txh.To, microuser.UserField_NICKNAME)

	if err != nil {
		logger.Info("get uid:", txh.To, " nick name err,", err)
	}
	penaltyTradeNo := GenerateTradeNo(constants.TRADE_TYPE_LIQUIDATED_DAMAGES,
		constants.TX_SUB_TYPE_LOCKD_LIQUIDATED_DAMAGES)
	trade := TradeInfo{
		TradeNo: penaltyTradeNo, Txid: txid, Status: constants.TRADE_STATUS_SUCC,
		Type:    constants.TRADE_TYPE_LIQUIDATED_DAMAGES,
		SubType: constants.TX_SUB_TYPE_LOCKD_LIQUIDATED_DAMAGES,
		From:    txh.From, To: txh.To, FromName: fromName, ToName: toName,
		Decimal: constants.TRADE_DECIMAIL, Amount: txh.Value,
		Currency: constants.TRADE_CURRENCY_LVTC, CreateTime: txh.Ts, FinishTime: txh.Ts,
	}
	err = InsertTradeInfo(trade)
	if err != nil {
		tx.Rollback()
		logger.Error("insert mongo db:dt_trades error ", err.Error())
		return false, constants.TRANS_ERR_SYS
	}
	err = tx.Commit()
	if err != nil {
		logger.Error("mysql commit  error ", err.Error())
		return false, constants.TRANS_ERR_SYS
	}

	return true, constants.TRANS_ERR_SUCC
}

func QuerySumLockAsset(uid int64, month int) int64 {
	row, err := gDBAsset.QueryRow("select if(sum(value) is null,0,sum(value)) as value from user_asset_lock where uid = ? and month >= ?", uid, month)
	if err != nil {
		logger.Error("query user asset lock error", err.Error())
		return 0
	}
	return utils.Str2Int64(row["value"])
}

func QuerySumLockAssetLvtc(uid int64, month int, currency string) int64 {
	row, err := gDBAsset.QueryRow("select if(sum(value) is null,0,sum(value)) as value from user_asset_lock where uid = ? and currency = ? and month >= ?", uid, currency, month)
	if err != nil {
		logger.Error("query user lvtc asset lock error", err.Error())
		return 0
	}
	return utils.Str2Int64(row["value"])
}

func QueryHashRateByUid(uid int64) int {

	sql := `select if(sum(t.h) is null,0,sum(t.h)) as sh from (
				select max(uh1.hashrate) as h from user_hashrate as uh1 where uh1.uid = ? and uh1.end = 0 group by uh1.type
				union all
				select uh2.hashrate as h from user_hashrate as uh2 where uh2.uid = ? and uh2.end >= ?
			) as t
			`

	params := []interface{}{
		uid,
		uid,
		utils.GetTimestamp13(),
	}

	row, err := gDBAsset.QueryRow(sql, params...)

	if err != nil || row == nil {
		if err != nil {
			logger.Error("query user asset lock error", err.Error())
		}
		return 0
	}
	return utils.Str2Int(row["sh"])
}

func QueryCountMinerByUid(uid int64) int {
	row, err := gDBAsset.QueryRow("select days from user_reward where uid = ?", uid)
	if err != nil {
		logger.Error("query reward days error", err.Error())
		return 0
	}
	return utils.Str2Int(row["days"])
}

func QueryLvtcCountMinerByUid(uid int64) int {
	row, err := gDBAsset.QueryRow("select days from user_reward_lvtc where uid = ?", uid)
	if err != nil {
		logger.Error("query reward days error", err.Error())
		return 0
	}
	if row == nil {
		// no record
		// query table: user_reward
		return QueryCountMinerByUid(uid)
	}
	return utils.Str2Int(row["days"])
}

func checkHashrateExists(uid int64, hrType int) bool {
	row, err := gDBAsset.QueryRow("select count(1) as c from user_hashrate where uid = ? and type = ? and end >= ?", uid, hrType, utils.GetTimestamp13())
	if err != nil {
		return true
	}
	if utils.Str2Int(row["c"]) > 0 {
		return true
	}
	return false
}

func GetUserWithdrawalQuotaByUid(uid int64) *UserWithdrawalQuota {
	row, err := gDBAsset.QueryRow("SELECT uid,`day`,`month`,casual,day_expend,last_level FROM user_withdrawal_quota where uid = ?", uid)

	if err != nil {
		logger.Error("query user withdraw quota error", err.Error())
		return nil
	}

	if row == nil {
		return nil
	}

	return convUserWithdrawalQuota(row)
}

func convUserWithdrawalQuota(al map[string]string) *UserWithdrawalQuota {
	if al == nil {
		return nil
	}
	alres := UserWithdrawalQuota{
		Day:       utils.Str2Int64(al["day"]),
		Month:     utils.Str2Int64(al["month"]),
		Casual:    utils.Str2Int64(al["casual"]),
		DayExpend: utils.Str2Int64(al["day_expend"]),
		LastLevel: utils.Str2Int(al["last_level"]),
	}
	return &alres
}

func CreateUserWithdrawalQuotaByTx(uid int64, day int64, month int64, lastLevel int, tx *sql.Tx) (sql.Result, error) {
	if tx == nil {
		tx, _ = gDBAsset.Begin()
		defer tx.Commit()
	}
	sql := "insert ignore into user_withdrawal_quota(uid, `day`, `month`, casual, day_expend, last_expend, last_income, last_level) values(?, ?, ?, ?, ?, ?, ?, ?) "
	return tx.Exec(sql, uid, day, month, 0, 0, utils.GetTimestamp13(), 0, lastLevel)
}

func ResetDayQuota(uid int64, dayQuota int64) bool {
	sql := "update user_withdrawal_quota set `day` = ? where uid = ?"
	result, err := gDBAsset.Exec(sql, dayQuota, uid)
	if err != nil {
		logger.Error("重置月额度错误" + err.Error())
		return false
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		return true
	} else {
		return false
	}
}

func ResetMonthQuota(uid int64, monthQuota int64, dayQuota int64, level int) bool {
	sql := "update user_withdrawal_quota set `month` = ?, `day` = ?, last_level = ? where uid = ?"
	result, err := gDBAsset.Exec(sql, monthQuota, dayQuota, level, uid)
	if err != nil {
		logger.Error("重置月额度错误" + err.Error())
		return false
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		return true
	} else {
		return false
	}
}

func UpdateLastLevelOfQuota(uid int64, lastLevel int) bool {
	sql := "update user_withdrawal_quota set last_level = ? where uid = ?"
	result, err := gDBAsset.Exec(sql, lastLevel, uid)
	if err != nil {
		logger.Error("更新提币额度last level错误", err.Error())
		return false
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		return true
	} else {
		return false
	}
}

func ExpendUserWithdrawalQuota(uid int64, expendQuota int64, quotaType int, tx *sql.Tx) (bool, error) {
	if expendQuota <= 0 {
		return false, errors.New("expend quota must greater than 0")
	}

	if quotaType != DAY_QUOTA_TYPE && quotaType != CASUAL_QUOTA_TYPE {
		return false, errors.New("expend quota type error")
	}

	if quotaType == DAY_QUOTA_TYPE {
		sql := "update user_withdrawal_quota set day = day - ?,month = month - ?,day_expend = ?,last_expend = ? where uid = ? and day >= ? and month >= ?"
		result, err := tx.Exec(sql, expendQuota, expendQuota, utils.GetTimestamp13(), utils.GetTimestamp13(), uid, expendQuota, expendQuota)
		if err != nil {
			logger.Error("expend day quota error ", err.Error())
			return false, err
		}
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected > 0 {
			return true, nil
		} else {
			logger.Info("update user withdrawal quota record of day quota miss")
			return false, nil
		}
	}
	if quotaType == CASUAL_QUOTA_TYPE {
		sql := "update user_withdrawal_quota set casual = casual - ?, month = month - ?, last_expend = ? where uid = ? and casual >= ? and month >= ?"
		result, err := tx.Exec(sql, expendQuota, expendQuota, utils.GetTimestamp13(), uid, expendQuota, expendQuota)
		if err != nil {
			logger.Error("expend casual quota error ", err.Error())
			return false, err
		}
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected > 0 {
			return true, nil
		} else {
			logger.Info("update user withdrawal quota record of casual miss")
			return false, nil
		}
	}
	return false, errors.New("record not exist")

}

func IncomeUserWithdrawalCasualQuota(uid int64, incomeCasual int64) (bool, error) {
	return IncomeUserWithdrawalCasualQuotaByTx(uid, incomeCasual, nil)
}

func IncomeUserWithdrawalCasualQuotaByTx(uid int64, incomeCasual int64, tx *sql.Tx) (bool, error) {
	if incomeCasual > 0 {
		if tx == nil {
			tx, _ = gDBAsset.Begin()
			defer tx.Commit()
		}
		sql := "update user_withdrawal_quota set casual = casual + ?,last_income = ? where uid = ?"
		result, err := tx.Exec(sql, incomeCasual, utils.GetTimestamp13(), uid)
		if err != nil {
			logger.Error("exec sql error", sql)
			return false, err
		}
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			return false, sqlBase.ErrNoRows
		}
	}
	return true, nil
}

func InitUserWithdrawal(uid int64) *UserWithdrawalQuota {
	return InitUserWithdrawalByTx(uid, nil)
}

func InitUserWithdrawalByTx(uid int64, tx *sql.Tx) *UserWithdrawalQuota {
	if tx == nil {
		tx, _ := gDBAsset.Begin()
		defer tx.Commit()
	}
	level := GetTransUserLevel(uid)
	limitConfig := config.GetLimitByLevel(level)
	day, month := limitConfig.DailyWithdrawalQuota()*utils.CONV_LVT,
		limitConfig.MonthlyWithdrawalQuota()*utils.CONV_LVT
	_, err := CreateUserWithdrawalQuotaByTx(uid, day, month, level, tx)
	if err != nil {
		logger.Error("insert user withdrawal quota error for user:", uid)
		return nil
	}
	return &UserWithdrawalQuota{
		Day:       day,
		Month:     month,
		Casual:    0,
		DayExpend: 0,
	}
}

//func Withdraw(uid int64, amount string, address string, currency string) (string, constants.Error) {
//	row, err := gDBAsset.QueryRow("select count(1) count from user_withdrawal_request where uid = ? and status in (?, ?, ?)", uid, constants.USER_WITHDRAWAL_REQUEST_WAIT_SEND, constants.USER_WITHDRAWAL_REQUEST_SEND, constants.USER_WITHDRAWAL_REQUEST_UNKNOWN)
//
//	if err != nil {
//		logger.Error("count processing reqeust from user_withdrawal_request error, error:", err.Error())
//		return "", constants.RC_SYSTEM_ERR
//	}
//
//	if utils.Str2Int(row["count"]) > 0 {
//		return "", constants.RC_HAS_UNFINISHED_WITHDRAWAL_TASK
//	}
//
//	tradeNo := GenerateTradeNo(constants.TRADE_TYPE_WITHDRAWAL, constants.TX_SUB_TYPE_WITHDRAW)
//
//	error := constants.RC_OK
//	switch currency {
//	case "LVTC":
//		error = withdrawLVTC(uid, amount, address, tradeNo)
//	case "ETH":
//		error = withdrawETH(uid, amount, address, tradeNo)
//	default:
//		error = constants.RC_INVALID_CURRENCY
//	}
//
//	return tradeNo, error
//
//}

func Withdraw(uid int64, amount, address, currency, feeCurrency, remark string, currencyDecimal, feeCurrencyDecimal int) (string, constants.Error) {
	CheckAndInitAsset(uid)
	//CheckAndInitAsset(config.GetWithdrawalConfig().WithdrawalAcceptAccount)
	//CheckAndInitAsset(config.GetWithdrawalConfig().FeeAcceptAccount)
	sql := "select count(1) count from user_withdrawal_request where uid = ?  and status in (?, ?, ?) and currency in (?"
	coins := config.GetConfig().GetChainCoinsBycoin(currency)
	if len(coins) == 0 {
		return "", constants.RC_PARAM_ERR
	}
	params := []interface{}{uid, constants.USER_WITHDRAWAL_REQUEST_WAIT_SEND, constants.USER_WITHDRAWAL_REQUEST_SEND, constants.USER_WITHDRAWAL_REQUEST_UNKNOWN}
	params = append(params, coins[0])
	for i := 1; i < len(coins); i++ {
		sql += ", ?"
		params = append(params, coins[i])
	}
	sql += ")"
	row, err := gDBAsset.QueryRow(sql, params...)
	if err != nil {
		logger.Error("count processing reqeust from user_withdrawal_request error, error:", err.Error())
		return "", constants.RC_SYSTEM_ERR
	}
	if utils.Str2Int(row["count"]) > 0 {
		return "", constants.RC_HAS_UNFINISHED_WITHDRAWAL_TASK
	}

	tradeNo := GenerateTradeNo(constants.TRADE_TYPE_WITHDRAWAL, constants.TX_SUB_TYPE_WITHDRAW)
	feeTradeNo := GenerateTradeNo(constants.TRADE_TYPE_FEE, constants.TX_SUB_TYPE_WITHDRAW_FEE)
	timestamp := utils.GetTimestamp13()
	fee, error := calculationFeeAndCheckQuotaForWithdraw(uid, amount, currency, feeCurrency, currencyDecimal)
	if error.Rc != constants.RC_OK.Rc {
		return "", error
	}

	feeInt := decimal.NewFromFloat(fee).Mul(decimal.NewFromFloat(float64(feeCurrencyDecimal))).IntPart()
	amountInt := utils.FloatStr2CoinsInt(amount, int64(currencyDecimal))

	tx, _ := gDBAsset.Begin()

	txId := GenerateTxID()
	txIdFee := GenerateTxID()
	//扣除提币资产
	error = transfer(txId, uid, config.GetWithdrawalConfig().WithdrawalAcceptAccount, amountInt, timestamp, currency, tradeNo, constants.TX_SUB_TYPE_WITHDRAW, tx)
	if error.Rc != constants.RC_OK.Rc {
		tx.Rollback()
		return "", error
	}
	//扣除手续费资产
	error = transfer(txIdFee, uid, config.GetWithdrawalConfig().FeeAcceptAccount, feeInt, timestamp, feeCurrency, feeTradeNo, constants.TX_SUB_TYPE_WITHDRAW_FEE, tx)
	if error.Rc != constants.RC_OK.Rc {
		tx.Rollback()
		return "", error
	}

	sql = "insert into user_withdrawal_request (trade_no, uid, value, address, txid, txid_fee, create_time, update_time, status, fee, currency, fee_currency, remark) values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	_, err = tx.Exec(sql, tradeNo, uid, amountInt, address, txId, txIdFee, timestamp, timestamp, constants.USER_WITHDRAWAL_REQUEST_WAIT_SEND, feeInt, currency, feeCurrency, remark)
	if err != nil {
		logger.Error("add user_withdrawal_request error ", err.Error())
		tx.Rollback()
		return "", constants.RC_SYSTEM_ERR
	}
	tx.Commit()

	go func() {
		var tradesArray []TradeInfo
		fromName, _ := rpc.GetUserField(uid, microuser.UserField_NICKNAME)
		toName, _ := rpc.GetUserField(config.GetWithdrawalConfig().FeeAcceptAccount, microuser.UserField_NICKNAME)
		feeCurrencyDecimal = 8
		if strings.EqualFold(feeCurrency, CURRENCY_EOS) {
			feeCurrencyDecimal = 4
		}
		feeTradeInfo := TradeInfo{
			TradeNo:         feeTradeNo,
			OriginalTradeNo: tradeNo,
			Type:            constants.TRADE_TYPE_FEE,
			SubType:         constants.TX_SUB_TYPE_WITHDRAW_FEE,
			From:            uid,
			FromName:        fromName,
			To:              config.GetWithdrawalConfig().FeeAcceptAccount,
			ToName:          toName,
			Amount:          feeInt,
			Decimal:         feeCurrencyDecimal,
			Currency:        feeCurrency,
			CreateTime:      timestamp,
			FinishTime:      timestamp,
			Status:          constants.TRADE_STATUS_SUCC,
			Txid:            txId,
		}
		tradesArray = append(tradesArray, feeTradeInfo)

		withdraw := TradeWithdrawal{
			Address: address,
		}
		toName, _ = rpc.GetUserField(config.GetWithdrawalConfig().WithdrawalAcceptAccount, microuser.UserField_NICKNAME)
		currencyDecimal = 8
		if strings.EqualFold(currency, CURRENCY_EOS) {
			currencyDecimal = 4
		}
		tradeInfo := TradeInfo{
			TradeNo:    tradeNo,
			Type:       constants.TRADE_TYPE_WITHDRAWAL,
			SubType:    constants.TX_SUB_TYPE_WITHDRAW,
			Subject:    remark,
			From:       uid,
			FromName:   fromName,
			To:         config.GetWithdrawalConfig().WithdrawalAcceptAccount,
			ToName:     toName,
			Amount:     amountInt,
			Decimal:    currencyDecimal,
			Currency:   currency,
			CreateTime: timestamp,
			Status:     constants.TRADE_STATUS_PROCESSING,
			Txid:       txId,
			FeeTradeNo: feeTradeNo,
			Withdrawal: &withdraw,
		}
		tradesArray = append(tradesArray, tradeInfo)
		err = InsertTradeInfo(tradesArray...)
		if err != nil {
			logger.Error("withdraw insert trade database error, error:", err.Error())
		}

		if strings.EqualFold(currency, CURRENCY_LVTC) {
			txh := &DTTXHistory{
				Id:       txId,
				TradeNo:  tradeNo,
				Type:     constants.TX_SUB_TYPE_WITHDRAW,
				From:     uid,
				To:       config.GetWithdrawalConfig().WithdrawalAcceptAccount,
				Value:    feeInt,
				Ts:       timestamp,
				Currency: CURRENCY_LVTC,
			}
			err := InsertLVTCCommited(txh)
			if err != nil {
				logger.Error("tx_history_lv_tmp insert mongo error ", err.Error())
				rdsDo("rpush", constants.PUSH_TX_HISTORY_LVT_QUEUE_NAME, utils.ToJSON(txh))
			} else {
				DeleteTxhistoryLvtTmpByTxid(txId)
			}
		}
	}()

	return tradeNo, error
}

func initBalanceInfoByTbName(tbName string, uid int64, tx *sql.Tx) error {
	sql := fmt.Sprintf("insert ignore into %s (uid,lastmodify) values (?,?) ", tbName)
	_, err := tx.Exec(sql, uid, utils.GetTimestamp13())
	return err
}

func transfer(txId, from, to, amount, timestamp int64, currency, tradeNo string, tradeType int, tx *sql.Tx) constants.Error {
	assetTableName := ""
	historyTableName := ""
	switch strings.ToUpper(currency) {
	case CURRENCY_BTC:
		assetTableName = "user_asset_btc"
		historyTableName = "tx_history_btc"
	case CURRENCY_ETH:
		assetTableName = "user_asset_eth"
		historyTableName = "tx_history_eth"
	case CURRENCY_EOS:
		assetTableName = "user_asset_eos"
		historyTableName = "tx_history_eos"
	case CURRENCY_LVTC:
		assetTableName = "user_asset_lvtc"
		historyTableName = "tx_history_lvt_tmp"
	default:
		logger.Error("currency not supported, currency:", currency)
		return constants.RC_PARAM_ERR
	}
	sql := fmt.Sprintf("select * from %s where uid in (?, ?) for update", assetTableName)
	_, err := tx.Exec(sql, from, to)
	if err != nil {
		logger.Error("lock ", currency, " asset error, error:", err.Error())
		//tx.Rollback()
		return constants.RC_SYSTEM_ERR
	}

	sql = fmt.Sprintf("select balance from %s where uid = ?", assetTableName)
	row := tx.QueryRow(sql, from)
	balance := int64(0)
	if err := row.Scan(&balance); err != nil {
		logger.Error("get balance err, uid:", from, " coin:", currency, "error:", err)
		return constants.RC_SYSTEM_ERR
	}
	if balance-amount < 0 {
		return constants.RC_INSUFFICIENT_BALANCE
	}

	if !checkAssetLimeted(assetTableName, from, tx) {
		//tx.Rollback()
		return constants.RC_ACCOUNT_ACCESS_LIMITED
	}

	sql = fmt.Sprintf("update %s set balance = balance - ?,lastmodify = ? where uid = ?", assetTableName)
	//扣除转出方balance
	info, err := tx.Exec(sql, amount, timestamp, from)
	if err != nil {
		logger.Error(strings.Contains(err.Error(), "1690"))
		logger.Error("exec sql(", sql, ") error ", err.Error())
		return constants.RC_SYSTEM_ERR
	}
	//update 以后校验修改记录条数，如果为0 说明初始化部分出现问题，返回错误
	rsa, _ := info.RowsAffected()
	if rsa == 0 {
		logger.Error("update user", currency, " balance error RowsAffected ", rsa, " can not find user  ", from, "")
		return constants.RC_PARAM_ERR
	}

	sql = fmt.Sprintf("update %s set balance = balance + ?,lastmodify = ? where uid = ?", assetTableName)
	//增加目标的balance
	info, err = tx.Exec(sql, amount, timestamp, to)
	if err != nil {
		logger.Error("exec sql(", sql, ") error ", err.Error())
		return constants.RC_SYSTEM_ERR
	}

	err = initBalanceInfoByTbName(assetTableName, to, tx)
	if err != nil {
		logger.Error("init user asset", assetTableName, " balance error RowsAffected ", to)
		return constants.RC_SYSTEM_ERR
	}

	//update 以后校验修改记录条数，如果为0 说明初始化部分出现问题，返回错误
	rsa, _ = info.RowsAffected()
	if rsa == 0 {
		logger.Error("update user", currency, " balance error RowsAffected ", rsa, " can not find user  ", to)
		return constants.RC_PARAM_ERR
	}

	sql = fmt.Sprintf("insert into %s (txid,type,trade_no,`from`,`to`,`value`,ts) values (?,?,?,?,?,?,?)", historyTableName)
	info, err = tx.Exec(sql, txId, tradeType, tradeNo, from, to, amount, timestamp)
	if err != nil {
		logger.Error("exec sql(", sql, ") error ", err.Error())
		return constants.RC_SYSTEM_ERR
	}
	rsa, _ = info.RowsAffected()
	if rsa == 0 {
		logger.Error("insert eth tx history failed, table:", historyTableName)
		return constants.RC_PARAM_ERR
	}

	if (strings.EqualFold(currency, constants.TRADE_CURRENCY_LVT) || strings.EqualFold(currency, CURRENCY_LVTC)) && tradeType == constants.TX_SUB_TYPE_TRANS {
		//txid 写入数据库
		_, e := InsertTXID(txId, tx)
		if e != nil {
			logger.Error("sql error ", e.Error())
			return constants.RC_SYSTEM_ERR
		}
	}
	return constants.RC_OK
}

func withdrawETH(uid int64, amount string, address, tradeNo string) constants.Error {
	timestamp := utils.GetTimestamp13()
	toETH := config.GetWithdrawalConfig().WithdrawalAcceptAccount
	amountInt := utils.FloatStrToLVTint(amount)
	feeToETH := config.GetWithdrawalConfig().FeeAcceptAccount
	feeTradeNo := GenerateTradeNo(constants.TRADE_TYPE_FEE, constants.TX_SUB_TYPE_WITHDRAW_FEE)
	txId := GenerateTxID()
	txIdFee := GenerateTxID()
	ethFee, error := calculationFeeAndCheckQuotaForWithdraw(uid, amount, constants.TRADE_CURRENCY_ETH, constants.TRADE_CURRENCY_ETH, CONV_LVT)
	if error.Rc != constants.RC_OK.Rc {
		return error
	}
	ethFeeInt := decimal.NewFromFloat(ethFee).Mul(decimal.NewFromFloat(CONV_LVT)).IntPart()
	ts := utils.GetTimestamp13()
	tx, _ := gDBAsset.Begin()
	_, err := tx.Exec("select * from user_asset_eth where uid in (?, ?, ?) for update", uid, toETH, feeToETH)
	if err != nil {
		logger.Error("lock eth asset error, error:", err.Error())
		tx.Rollback()
		return constants.RC_SYSTEM_ERR
	}
	//查询转出账户余额是否满足需要 使用新的校验方法，考虑到锁仓的问题
	if !ckeckEthBalance(uid, amountInt+ethFeeInt, tx) {
		tx.Rollback()
		return constants.RC_INSUFFICIENT_BALANCE
	}

	error = ethTransfer(txId, uid, toETH, amountInt, ts, tradeNo, constants.TX_SUB_TYPE_WITHDRAW, tx)
	if error.Rc != constants.RC_OK.Rc {
		tx.Rollback()
		return error
	}

	error = ethTransfer(txIdFee, uid, feeToETH, ethFeeInt, ts, feeTradeNo, constants.TX_SUB_TYPE_WITHDRAW_FEE, tx)
	if error.Rc != constants.RC_OK.Rc {
		tx.Rollback()
		return error
	}

	sql := "insert into user_withdrawal_request (trade_no, uid, value, address, txid, txid_fee, create_time, update_time, status, fee, currency) values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	_, err = tx.Exec(sql, tradeNo, uid, amountInt, address, txId, txIdFee, timestamp, timestamp, constants.USER_WITHDRAWAL_REQUEST_WAIT_SEND, ethFeeInt, constants.TRADE_CURRENCY_ETH)
	if err != nil {
		logger.Error("add user_withdrawal_request error ", err.Error())
		tx.Rollback()
		return constants.RC_SYSTEM_ERR
	}
	tx.Commit()

	go func() {
		err = addFeeTradeInfo(txIdFee, feeTradeNo, tradeNo, constants.TRADE_TYPE_FEE, constants.TX_SUB_TYPE_WITHDRAW_FEE, uid, feeToETH, ethFeeInt, constants.TRADE_CURRENCY_ETH, 8, timestamp)
		if err != nil {
			logger.Error("withdraw fee insert trade database error, error:", err.Error())
		}
		err = addTradeInfo(txId, tradeNo, constants.TRADE_TYPE_WITHDRAWAL, constants.TX_SUB_TYPE_WITHDRAW, uid, toETH, address, amountInt, constants.TRADE_CURRENCY_ETH, feeTradeNo, 8, timestamp)
		if err != nil {
			logger.Error("withdraw insert trade database error, error:", err.Error())
		}
	}()

	return constants.RC_OK
}

func ethTransfer(txId, from, to, amount, timestamp int64, tradeNo string, tradeType int, tx *sql.Tx) constants.Error {
	//扣除提币手续费
	//扣除转出方balance
	info, err := tx.Exec("update user_asset_eth set balance = balance - ?,lastmodify = ? where uid = ?", amount, timestamp, from)
	if err != nil {
		logger.Error("sql error ", err.Error())
		return constants.RC_SYSTEM_ERR
	}
	//update 以后校验修改记录条数，如果为0 说明初始化部分出现问题，返回错误
	rsa, _ := info.RowsAffected()
	if rsa == 0 {
		logger.Error("update user balance error RowsAffected ", rsa, " can not find user  ", from, "")
		return constants.RC_PARAM_ERR
	}

	//增加目标的balance
	info, err = tx.Exec("update user_asset_eth set balance = balance + ?,lastmodify = ? where uid = ?", amount, timestamp, to)
	if err != nil {
		logger.Error("sql error ", err.Error())
		return constants.RC_SYSTEM_ERR
	}
	//update 以后校验修改记录条数，如果为0 说明初始化部分出现问题，返回错误
	rsa, _ = info.RowsAffected()
	if rsa == 0 {
		logger.Error("update user balance error RowsAffected ", rsa, " can not find user  ", to, "")
		return constants.RC_PARAM_ERR
	}

	info, err = tx.Exec("insert into tx_history_eth (txid,type,trade_no,`from`,`to`,`value`,ts) values (?,?,?,?,?,?,?)",
		txId, tradeType, tradeNo, from, to, amount, timestamp,
	)
	if err != nil {
		logger.Error("sql error ", err.Error())
		return constants.RC_SYSTEM_ERR
	}
	rsa, _ = info.RowsAffected()
	if rsa == 0 {
		logger.Error("insert eth tx history failed")
		return constants.RC_PARAM_ERR
	}
	return constants.RC_OK
}

func withdrawLVTC(uid int64, amount string, address, tradeNo string) constants.Error {
	timestamp := utils.GetTimestamp13()
	txId := GenerateTxID()
	toLvt := config.GetWithdrawalConfig().WithdrawalAcceptAccount
	amountInt := utils.FloatStrToLVTint(amount)
	ethFee, err := calculationFeeAndCheckQuotaForWithdraw(uid, amount, constants.TRADE_CURRENCY_LVTC, constants.TRADE_CURRENCY_ETH, CONV_LVT)
	if err.Rc != constants.RC_OK.Rc {
		return err
	}

	tx, _ := gDBAsset.Begin()
	tx.Exec("select * from user_withdrawal_request where uid = ? for update", uid)

	ethFeeInt := decimal.NewFromFloat(ethFee).Mul(decimal.NewFromFloat(CONV_LVT)).IntPart()

	transLvtResult, e := TransAccountLvtcByTx(txId, uid, toLvt, amountInt, tx)
	if !transLvtResult {
		tx.Rollback()
		switch e {
		case constants.TRANS_ERR_INSUFFICIENT_BALANCE:
			return constants.RC_INSUFFICIENT_BALANCE
		case constants.TRANS_ERR_SYS:
			return constants.RC_TRANS_IN_PROGRESS
		case constants.TRANS_ERR_ASSET_LIMITED:
			return constants.RC_ACCOUNT_ACCESS_LIMITED
		default:
			return constants.RC_SYSTEM_ERR
		}
	}
	_, err3 := tx.Exec("insert into tx_history_lvt_tmp (txid, type, trade_no, `from`, `to`, value, ts) VALUES (?, ?, ?, ?, ?, ?, ?)", txId, constants.TX_SUB_TYPE_WITHDRAW, tradeNo, uid, toLvt, amountInt, timestamp)
	if err3 != nil {
		tx.Rollback()
		logger.Error("insert tx_history_lvt_tmp error ", err3.Error())
		return constants.RC_SYSTEM_ERR
	}

	toEth := config.GetWithdrawalConfig().FeeAcceptAccount
	feeTradeNo := GenerateTradeNo(constants.TRADE_TYPE_FEE, constants.TX_SUB_TYPE_WITHDRAW_FEE)
	txIdFee, e := EthTransCommit(-1, uid, toEth, ethFeeInt, feeTradeNo, constants.TX_SUB_TYPE_WITHDRAW_FEE, tx)
	if txIdFee <= 0 {
		tx.Rollback()
		switch e {
		case constants.TRANS_ERR_INSUFFICIENT_BALANCE:
			return constants.RC_INSUFFICIENT_BALANCE
		case constants.TRANS_ERR_SYS:
			return constants.RC_TRANS_IN_PROGRESS
		case constants.TRANS_ERR_ASSET_LIMITED:
			return constants.RC_ACCOUNT_ACCESS_LIMITED
		default:
			return constants.RC_SYSTEM_ERR
		}
	}

	sql := "insert into user_withdrawal_request (trade_no, uid, value, address, txid, txid_fee, create_time, update_time, status, fee, currency) values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	_, err1 := tx.Exec(sql, tradeNo, uid, amountInt, address, txId, txIdFee, timestamp, timestamp, constants.USER_WITHDRAWAL_REQUEST_WAIT_SEND, ethFeeInt, constants.TRADE_CURRENCY_LVTC)
	if err1 != nil {
		logger.Error("add user_withdrawal_request error ", err1.Error())
		tx.Rollback()
		return constants.RC_SYSTEM_ERR
	}
	tx.Commit()

	//同步至mongo
	go func() {
		txh := &DTTXHistory{
			Id:       txId,
			TradeNo:  tradeNo,
			Type:     constants.TX_SUB_TYPE_WITHDRAW,
			From:     uid,
			To:       toLvt,
			Value:    amountInt,
			Ts:       timestamp,
			Currency: "LVTC",
		}
		err := InsertLVTCCommited(txh)
		if err != nil {
			logger.Error("tx_history_lv_tmp insert mongo error ", err.Error())
			rdsDo("rpush", constants.PUSH_TX_HISTORY_LVT_QUEUE_NAME, utils.ToJSON(txh))
		} else {
			DeleteTxhistoryLvtTmpByTxid(txId)
		}

		err = addFeeTradeInfo(txIdFee, feeTradeNo, tradeNo, constants.TRADE_TYPE_FEE, constants.TX_SUB_TYPE_WITHDRAW_FEE, uid, toEth, ethFeeInt, constants.TRADE_CURRENCY_ETH, 8, timestamp)
		if err != nil {
			logger.Error("withdraw fee insert trade database error, error:", err.Error())
		}
		err = addTradeInfo(txId, tradeNo, constants.TRADE_TYPE_WITHDRAWAL, constants.TX_SUB_TYPE_WITHDRAW, uid, toLvt, address, amountInt, constants.TRADE_CURRENCY_LVTC, feeTradeNo, 8, timestamp)
		if err != nil {
			logger.Error("withdraw insert trade database error, error:", err.Error())
		}
	}()
	return constants.RC_OK
}

func calculationFeeAndCheckQuotaForWithdraw(uid int64, withdrawAmount, currency, feeCurrency string, currencyDecimal int) (float64, constants.Error) {
	if utils.Str2Float64(withdrawAmount) <= 0 {
		return float64(0), constants.RC_PARAM_ERR
	}
	withdrawQuota := getWithdrawQuota(currency)
	if withdrawQuota == nil {
		return float64(0), constants.RC_PARAM_ERR
	}
	if withdrawQuota.SingleAmountMin > 0 && withdrawQuota.SingleAmountMin > utils.Str2Float64(withdrawAmount) {
		return float64(0), constants.RC_TRANS_AMOUNT_EXCEEDING_LIMIT
	}

	if withdrawQuota.DailyAmountMax > 0 {
		sql := "select sum(value) total_value from user_withdrawal_request where uid = ? and currency = ? and status in (?, ?, ?, ?) and create_time >= ?"
		row, err := gDBAsset.QueryRow(sql, uid, currency, constants.USER_WITHDRAWAL_REQUEST_WAIT_SEND, constants.USER_WITHDRAWAL_REQUEST_SEND, constants.USER_WITHDRAWAL_REQUEST_SUCCESS, constants.USER_WITHDRAWAL_REQUEST_UNKNOWN, utils.GetTimestamp13ByTime(utils.GetDayStart(utils.GetTimestamp13())))
		if err != nil {
			logger.Error("query that day total withdraw amount error, uid:", uid, ",error:", err.Error())
		}
		totalAmount := utils.Str2Int64(row["total_value"])

		dailyAmount, _ := decimal.NewFromString(utils.Float642Str(withdrawQuota.DailyAmountMax))
		dailyAmountInt64 := dailyAmount.Mul(decimal.NewFromFloat(float64(currencyDecimal))).IntPart()
		withdrawAmountBig, _ := decimal.NewFromString(withdrawAmount)
		withdrawAmountInt64 := withdrawAmountBig.Mul(decimal.NewFromFloat(float64(currencyDecimal))).IntPart()

		logger.Info("daily quota out limit, withdraw amount:", withdrawAmountInt64, ",that day withdraw total amount:", totalAmount, ",daily amount:", dailyAmountInt64)
		if dailyAmountInt64 < totalAmount+withdrawAmountInt64 {
			return float64(0), constants.RC_TRANS_AMOUNT_EXCEEDING_LIMIT
		}
	}
	var feeOfWithdraw = float64(-1)
	for _, fee := range withdrawQuota.Fee {
		if fee.FeeCurrency == feeCurrency {
			switch fee.FeeType {
			case 0:
				feeOfWithdraw = fee.FeeFixed
			case 1:
				feeOfWithdraw = utils.Str2Float64(withdrawAmount) * fee.FeeRate
				if feeOfWithdraw < fee.FeeMin {
					feeOfWithdraw = fee.FeeMin
				}
				if feeOfWithdraw > fee.FeeMax {
					feeOfWithdraw = fee.FeeMax
				}
			default:
				logger.Error("withdraw fee type not supported, fee currency:", feeCurrency, "fee type:", fee.FeeType)
				return float64(0), constants.RC_INVALID_CURRENCY
			}
		}
	}

	if feeOfWithdraw < 0 {
		logger.Error("withdraw fee err, fee currency:", feeCurrency, "fee:", feeOfWithdraw)
		return float64(0), constants.RC_TRANSFER_FEE_ERROR
	}

	return feeOfWithdraw, constants.RC_OK
}

func getWithdrawQuota(withdrawCurrency string) *WithdrawQuota {
	key := constants.WITHDRAW_QUOTA_FEE_KEY + "_" + withdrawCurrency
	withdrawQuota := new(WithdrawQuota)
	results, _ := redis.String(rdsDo("GET", key))
	reload := false
	if len(results) > 0 {
		if err := utils.FromJson(results, withdrawQuota); err != nil {
			logger.Error("get withdraw quota from redis error, error:", err.Error())
			reload = true
		}
	} else {
		reload = true
	}
	if reload {
		withdrawQuota = GetWithdrawQuotaByCurrency(withdrawCurrency)
		tomorrow, _ := time.ParseInLocation("2006-01-02", time.Now().Format("2006-01-02")+" 23:59:59", time.Local)
		rdsDo("SET", key, utils.ToJSON(withdrawQuota), "EX", tomorrow.Unix()+1-utils.GetTimestamp10())
	}
	return withdrawQuota
}

//status为成功
func addFeeTradeInfo(txid int64, tradeNo string, originalTradeNo string, tradeType, subType int, from int64, to int64, amount int64, currency string, currencyDecimal int, ts int64) error {
	fromName, _ := rpc.GetUserField(from, microuser.UserField_NICKNAME)
	toName, _ := rpc.GetUserField(to, microuser.UserField_NICKNAME)
	tradeInfo := TradeInfo{
		TradeNo:         tradeNo,
		OriginalTradeNo: originalTradeNo,
		Type:            tradeType,
		SubType:         subType,
		From:            from,
		FromName:        fromName,
		To:              to,
		ToName:          toName,
		Amount:          amount,
		Decimal:         currencyDecimal,
		Currency:        currency,
		CreateTime:      ts,
		FinishTime:      ts,
		Status:          constants.TRADE_STATUS_SUCC,
		Txid:            txid,
	}
	return InsertTradeInfo(tradeInfo)
}

//status为处理中
func addTradeInfo(txid int64, tradeNo string, tradeType, subType int, from int64, to int64, address string, amount int64, currency, FeeTradeNo string, currencyDecimal int, ts int64) error {
	withdraw := TradeWithdrawal{
		Address: address,
	}
	fromName, _ := rpc.GetUserField(from, microuser.UserField_NICKNAME)
	toName, _ := rpc.GetUserField(to, microuser.UserField_NICKNAME)
	tradeInfo := TradeInfo{
		TradeNo:    tradeNo,
		Type:       tradeType,
		SubType:    subType,
		From:       from,
		FromName:   fromName,
		To:         to,
		ToName:     toName,
		Amount:     amount,
		Decimal:    currencyDecimal,
		Currency:   currency,
		CreateTime: ts,
		Status:     1,
		Txid:       txid,
		FeeTradeNo: FeeTradeNo,
		Withdrawal: &withdraw,
	}
	return InsertTradeInfo(tradeInfo)
}

func DeleteTxhistoryLvtTmpByTxid(txid int64) {
	gDBAsset.Exec("delete from tx_history_lvt_tmp where txid= ? ", txid)
}

func QueryTxhistoryLvtTmpByTimie(ts int64) []*DTTXHistory {
	if gDBAsset != nil && gDBAsset.IsConn() {
		results := gDBAsset.Query("select thl.*,uwr.currency from tx_history_lvt_tmp thl left join user_withdrawal_request uwr on thl.trade_no = uwr.trade_no where ts < ?", ts)
		if results == nil {
			return nil
		}
		return convDTTXHistoryList(results)
	}
	return nil
}

func convDTTXHistoryList(list []map[string]string) []*DTTXHistory {
	listRes := make([]*DTTXHistory, 0)
	for _, v := range list {
		listRes = append(listRes, convDTTXHistoryRequest(v))
	}
	return listRes
}

func convDTTXHistoryRequest(al map[string]string) *DTTXHistory {
	if al == nil {
		return nil
	}
	alres := DTTXHistory{
		Id:       utils.Str2Int64(al["txid"]),
		Type:     utils.Str2Int(al["type"]),
		TradeNo:  al["trade_no"],
		From:     utils.Str2Int64(al["from"]),
		To:       utils.Str2Int64(al["to"]),
		Value:    utils.Str2Int64(al["value"]),
		Ts:       utils.Str2Int64(al["ts"]),
		Currency: al["currency"],
	}
	return &alres
}

func GetWithdrawalCurrency(tradeNo string) string {
	row, _ := gDBAsset.QueryRow("select currency from user_withdrawal_request where trade_no = ?", tradeNo)
	return row["currency"]
}

func QueryWithdrawalList(uid int64) []*UserWithdrawalRequest {
	sql := "select id, trade_no, currency, value, address, fee, create_time, update_time, case status when 0 then 1 else status end status from user_withdrawal_request where uid = ? order by create_time desc"
	results := gDBAsset.Query(sql, uid)
	if results == nil {
		return nil
	}
	return convUserWithdrawalRequestList(results)
}

func convUserWithdrawalRequestList(list []map[string]string) []*UserWithdrawalRequest {
	listRes := make([]*UserWithdrawalRequest, 0)
	for _, v := range list {
		listRes = append(listRes, convUserWithdrawalRequest(v))
	}
	return listRes
}

func convUserWithdrawalRequest(al map[string]string) *UserWithdrawalRequest {
	if al == nil {
		return nil
	}
	alres := UserWithdrawalRequest{
		Id:         utils.Str2Int64(al["id"]),
		TradeNo:    al["trade_no"],
		Address:    al["address"],
		Value:      utils.Str2Int64(al["value"]),
		Currency:   al["currency"],
		Fee:        utils.Str2Int64(al["fee"]),
		CreateTime: utils.Str2Int64(al["create_time"]),
		UpdateTime: utils.Str2Int64(al["update_time"]),
		Status:     utils.Str2Int(al["status"]),
	}
	return &alres
}

func BtcTransCommit(txid, from, to, value int64, tradeNo string, tradeType int, tx *sql.Tx) (int64, int) {
	if tx == nil {
		tx, _ = gDBAsset.Begin()
		defer tx.Commit()
	}

	tx.Exec("select * from user_asset_btc where uid in (?,?) for update", from, to)

	ts := utils.GetTimestamp13()

	//查询转出账户余额是否满足需要 使用新的校验方法，考虑到锁仓的问题
	if !ckeckBtcBalance(from, value, tx) {
		return 0, constants.TRANS_ERR_INSUFFICIENT_BALANCE
	}
	if !checkAssetLimeted("user_asset_btc", from, tx) {
		return 0, constants.TRANS_ERR_ASSET_LIMITED
	}

	//扣除转出方balance
	info1, err1 := tx.Exec("update user_asset_btc set balance = balance - ?,lastmodify = ? where uid = ?", value, ts, from)
	if err1 != nil {
		logger.Error("sql error ", err1.Error())
		return 0, constants.TRANS_ERR_SYS
	}
	//update 以后校验修改记录条数，如果为0 说明初始化部分出现问题，返回错误
	rsa, _ := info1.RowsAffected()
	if rsa == 0 {
		logger.Error("update user balance error RowsAffected ", rsa, " can not find user  ", from, "")
		return 0, constants.TRANS_ERR_SYS
	}

	//增加目标的balance
	info2, err2 := tx.Exec("update user_asset_btc set balance = balance + ?,lastmodify = ? where uid = ?", value, ts, to)
	if err2 != nil {
		logger.Error("sql error ", err2.Error())
		return 0, constants.TRANS_ERR_SYS
	}
	//update 以后校验修改记录条数，如果为0 说明初始化部分出现问题，返回错误
	rsa, _ = info2.RowsAffected()
	if rsa == 0 {
		logger.Error("update user balance error RowsAffected ", rsa, " can not find user  ", to, "")
		return 0, constants.TRANS_ERR_SYS
	}

	if txid < 0 {
		txid = GenerateTxID()
		if txid < 0 {
			logger.Error("can not get txid")
			return 0, constants.TRANS_ERR_SYS
		}
	}
	info3, err3 := tx.Exec("insert into tx_history_btc (txid,type,trade_no,`from`,`to`,`value`,ts) values (?,?,?,?,?,?,?)",
		txid, tradeType, tradeNo, from, to, value, ts,
	)
	if err3 != nil {
		logger.Error("sql error ", err3.Error())
		return 0, constants.TRANS_ERR_SYS
	}
	rsa, _ = info3.RowsAffected()
	if rsa == 0 {
		logger.Error("insert btc tx history failed")
		return 0, constants.TRANS_ERR_SYS
	}
	return txid, constants.TRANS_ERR_SUCC
}

func EosTransCommit(txid, from, to, value int64, tradeNo string, tradeType int, tx *sql.Tx) (int64, int) {
	if tx == nil {
		tx, _ = gDBAsset.Begin()
		defer tx.Commit()
	}

	tx.Exec("select * from user_asset_eos where uid in (?,?) for update", from, to)

	ts := utils.GetTimestamp13()

	//查询转出账户余额是否满足需要 使用新的校验方法，考虑到锁仓的问题
	if !ckeckEosBalance(from, value, tx) {
		return 0, constants.TRANS_ERR_INSUFFICIENT_BALANCE
	}
	if !checkAssetLimeted("user_asset_eos", from, tx) {
		return 0, constants.TRANS_ERR_ASSET_LIMITED
	}

	//扣除转出方balance
	info1, err1 := tx.Exec("update user_asset_eos set balance = balance - ?,lastmodify = ? where uid = ?", value, ts, from)
	if err1 != nil {
		logger.Error("sql error ", err1.Error())
		return 0, constants.TRANS_ERR_SYS
	}
	//update 以后校验修改记录条数，如果为0 说明初始化部分出现问题，返回错误
	rsa, _ := info1.RowsAffected()
	if rsa == 0 {
		logger.Error("update user balance error RowsAffected ", rsa, " can not find user  ", from, "")
		return 0, constants.TRANS_ERR_SYS
	}

	//增加目标的balance
	info2, err2 := tx.Exec("update user_asset_eos set balance = balance + ?,lastmodify = ? where uid = ?", value, ts, to)
	if err2 != nil {
		logger.Error("sql error ", err2.Error())
		return 0, constants.TRANS_ERR_SYS
	}
	//update 以后校验修改记录条数，如果为0 说明初始化部分出现问题，返回错误
	rsa, _ = info2.RowsAffected()
	if rsa == 0 {
		logger.Error("update user balance error RowsAffected ", rsa, " can not find user  ", to, "")
		return 0, constants.TRANS_ERR_SYS
	}

	if txid < 0 {
		txid = GenerateTxID()
		if txid < 0 {
			logger.Error("can not get txid")
			return 0, constants.TRANS_ERR_SYS
		}
	}
	info3, err3 := tx.Exec("insert into tx_history_eos (txid,type,trade_no,`from`,`to`,`value`,ts) values (?,?,?,?,?,?,?)",
		txid, tradeType, tradeNo, from, to, value, ts,
	)
	if err3 != nil {
		logger.Error("sql error ", err3.Error())
		return 0, constants.TRANS_ERR_SYS
	}
	rsa, _ = info3.RowsAffected()
	if rsa == 0 {
		logger.Error("insert eos tx history failed")
		return 0, constants.TRANS_ERR_SYS
	}
	return txid, constants.TRANS_ERR_SUCC
}

func EthTransCommit(txid, from, to, value int64, tradeNo string, tradeType int, tx *sql.Tx) (int64, int) {
	if tx == nil {
		tx, _ = gDBAsset.Begin()
		defer tx.Commit()
	}

	tx.Exec("select * from user_asset_eth where uid in (?,?) for update", from, to)

	ts := utils.GetTimestamp13()

	//查询转出账户余额是否满足需要 使用新的校验方法，考虑到锁仓的问题
	if !ckeckEthBalance(from, value, tx) {
		return 0, constants.TRANS_ERR_INSUFFICIENT_BALANCE
	}
	if !checkAssetLimeted("user_asset_eth", from, tx) {
		return 0, constants.TRANS_ERR_ASSET_LIMITED
	}

	//扣除转出方balance
	info1, err1 := tx.Exec("update user_asset_eth set balance = balance - ?,lastmodify = ? where uid = ?", value, ts, from)
	if err1 != nil {
		logger.Error("sql error ", err1.Error())
		return 0, constants.TRANS_ERR_SYS
	}
	//update 以后校验修改记录条数，如果为0 说明初始化部分出现问题，返回错误
	rsa, _ := info1.RowsAffected()
	if rsa == 0 {
		logger.Error("update user balance error RowsAffected ", rsa, " can not find user  ", from, "")
		return 0, constants.TRANS_ERR_SYS
	}

	//增加目标的balance
	info2, err2 := tx.Exec("update user_asset_eth set balance = balance + ?,lastmodify = ? where uid = ?", value, ts, to)
	if err2 != nil {
		logger.Error("sql error ", err2.Error())
		return 0, constants.TRANS_ERR_SYS
	}
	//update 以后校验修改记录条数，如果为0 说明初始化部分出现问题，返回错误
	rsa, _ = info2.RowsAffected()
	if rsa == 0 {
		logger.Error("update user balance error RowsAffected ", rsa, " can not find user  ", to, "")
		return 0, constants.TRANS_ERR_SYS
	}

	if txid < 0 {
		txid = GenerateTxID()
		if txid < 0 {
			logger.Error("can not get txid")
			return 0, constants.TRANS_ERR_SYS
		}
	}
	info3, err3 := tx.Exec("insert into tx_history_eth (txid,type,trade_no,`from`,`to`,`value`,ts) values (?,?,?,?,?,?,?)",
		txid, tradeType, tradeNo, from, to, value, ts,
	)
	if err3 != nil {
		logger.Error("sql error ", err3.Error())
		return 0, constants.TRANS_ERR_SYS
	}
	rsa, _ = info3.RowsAffected()
	if rsa == 0 {
		logger.Error("insert eth tx history failed")
		return 0, constants.TRANS_ERR_SYS
	}
	return txid, constants.TRANS_ERR_SUCC
}

func InsertTradePending(txid, from, to int64, tradeNo, bizContent string, value int64, tradeType int) error {
	tradeSql := "insert into trade_pending (txid,trade_no,`from`,`to`,type,biz_content,`value`,ts) values (?,?,?,?,?,?,?,?)"
	_, err := gDBAsset.Exec(tradeSql,
		txid, tradeNo, from, to, tradeType,
		bizContent, value, utils.GetTimestamp13())
	return err
}

func GetTradePendingByTxid(txid string, uid int64) (*TradePending, error) {
	sql := "select * from trade_pending where txid = ? and `from` = ?"
	row, err := gDBAsset.QueryRow(sql, txid, uid)
	if err != nil {
		logger.Error("query trade_pending error", err.Error())
		return nil, err
	}
	return ConvTradePending(row), nil
}

func ConvTradePending(row map[string]string) *TradePending {
	if row == nil {
		return nil
	}
	tp := new(TradePending)
	tp.Txid = row["txid"]
	tp.TradeNo = row["trade_no"]
	tp.BizContent = row["biz_content"]
	tp.From = utils.Str2Int64(row["from"])
	tp.To = utils.Str2Int64(row["to"])
	tp.Ts = utils.Str2Int64(row["ts"])
	tp.Type = utils.Str2Int(row["type"])
	tp.Value = utils.Str2Int64(row["value"])
	tp.ValueStr = utils.LVTintToFloatStr(tp.Value)
	return tp
}

func DeleteTradePending(tradeNo string, uid int64, tx *sql.Tx) error {
	if tx == nil {
		tx, _ = gDBAsset.Begin()
		defer tx.Commit()
	}
	_, err := tx.Exec("delete from trade_pending where trade_no = ? and `from` = ?", tradeNo, uid)
	return err
}

func InsertWithdrawalCardUse(wcu *UserWithdrawalCardUse) error {
	return InsertWithdrawalCardUseByTx(wcu, nil)
}
func InsertWithdrawalCardUseByTx(wcu *UserWithdrawalCardUse, tx *sql.Tx) error {
	if tx == nil {
		tx, _ = gDBAsset.Begin()
		defer tx.Commit()
	}
	_, err := tx.Exec("insert into user_withdrawal_card_use (trade_no,uid,quota,cost,create_time,type,currency,txid) values (?,?,?,?,?,?,?,?)",
		wcu.TradeNo,
		wcu.Uid,
		wcu.Quota,
		wcu.Cost,
		wcu.CreateTime,
		wcu.Type,
		wcu.Currency,
		wcu.Txid,
	)
	return err
}

func CheckEthPending(tradeNo string) bool {
	row, err := gDBAsset.QueryRow("select count(1) as c from trade_pending where trade_no = ?", tradeNo)
	if err != nil {
		logger.Error("query db error", err.Error())
		return false
	}
	if row == nil {
		return false
	}
	return utils.Str2Int(row["c"]) > 0
}

func CheckTradePendingByTxid(txid int64) bool {
	row, err := gDBAsset.QueryRow("select count(1) as c from trade_pending where txid = ?", txid)
	if err != nil {
		logger.Error("query db error", err.Error())
		return false
	}
	if row == nil {
		return false
	}
	return utils.Str2Int(row["c"]) > 0
}

func CheckEthHistory(tradeNo string) bool {
	row, err := gDBAsset.QueryRow("select count(1) as c from tx_history_eth where trade_no = ?", tradeNo)
	if err != nil {
		logger.Error("query db error", err.Error())
		return false
	}
	if row == nil {
		return false
	}
	return utils.Str2Int(row["c"]) > 0
}

func CheckEthHistoryByTxid(txid int64) bool {
	row, err := gDBAsset.QueryRow("select count(1) as c from tx_history_eth where txid = ?", txid)
	if err != nil {
		logger.Error("query db error", err.Error())
		return false
	}
	if row == nil {
		return false
	}
	return utils.Str2Int(row["c"]) > 0
}

func CheckEosHistoryByTxid(txid int64) bool {
	row, err := gDBAsset.QueryRow("select count(1) as c from tx_history_eos where txid = ?", txid)
	if err != nil {
		logger.Error("query db error", err.Error())
		return false
	}
	if row == nil {
		return false
	}
	return utils.Str2Int(row["c"]) > 0
}

func CheckBtcHistoryByTxid(txid int64) bool {
	row, err := gDBAsset.QueryRow("select count(1) as c from tx_history_btc where txid = ?", txid)
	if err != nil {
		logger.Error("query db error", err.Error())
		return false
	}
	if row == nil {
		return false
	}
	return utils.Str2Int(row["c"]) > 0
}

func QueryEthTxHistory(uid int64, txid string, tradeType int, begin, end int64, max int) []map[string]string {
	sql := "select * from tx_history_eth where `from` = ?"
	params := []interface{}{uid}
	if len(txid) > 0 {
		sql += " and txid = ?"
		params = append(params, utils.Str2Int64(txid))
	} else {
		if tradeType > 0 {
			sql += " and `type` = ?"
			params = append(params, tradeType)
		}
		if begin > 0 {
			sql += " and `ts` >= ?"
			params = append(params, begin)
		}
		if end > 0 {
			sql += " and `ts` <= ?"
			params = append(params, end)
		}
	}

	sql += " union select * from tx_history_eth where `to` = ?"
	params = append(params, uid)
	if len(txid) > 0 {
		sql += " and txid = ?"
		params = append(params, utils.Str2Int64(txid))
	} else {
		if tradeType > 0 {
			sql += " and `type` = ?"
			params = append(params, tradeType)
		}
		if begin > 0 {
			sql += " and `ts` >= ?"
			params = append(params, begin)
		}
		if end > 0 {
			sql += " and `ts` <= ?"
			params = append(params, end)
		}
	}
	sql += "  order by txid desc limit ?"
	params = append(params, max)
	rows := gDBAsset.Query(sql, params...)
	return rows
}

func GetUserWithdrawCardByUid(uid int64) []map[string]string {
	rows := gDBAsset.Query("select * from withdrawal_card where owner_uid = ? order by get_time desc", uid)
	return rows
}

func GetUserWithdrawCardUseByUid(uid int64) []map[string]string {
	rows := gDBAsset.Query("select * from user_withdrawal_card_use where uid = ? order by create_time desc", uid)
	return rows
}

func GetUserWithdrawCardByPwd(pwd string) *UserWithdrawCard {
	if row, err := gDBAsset.QueryRow("select * from withdrawal_card where password = ?", pwd); err != nil {
		logger.Error("get user withdraw card failed", err.Error())
		return nil
	} else {
		return convUserWithdrawCard(row)
	}
}

func UseWithdrawCard(card *UserWithdrawCard, uid int64) error {
	tradeNo := GenerateTradeNo(constants.TRADE_NO_BASE_TYPE, constants.TRADE_NO_TYPE_USE_COIN_CARD)
	ts := utils.GetTimestamp13()
	tx, err := gDBAsset.Begin()
	if err != nil {
		return err
	}

	tx.Exec("select * from withdrawal_card where id = ? for update", card.Id)

	_, err = tx.Exec("update withdrawal_card set status = ?,use_time = ? ,trade_no = ? where id = ?", constants.WITHDRAW_CARD_STATUS_USE, ts, tradeNo, card.Id)
	if err != nil {
		tx.Rollback()
		return err
	}

	wcu := &UserWithdrawalCardUse{
		TradeNo:    tradeNo,
		Uid:        uid,
		Quota:      card.Quota,
		Cost:       card.Cost,
		CreateTime: ts,
		Type:       constants.WITHDRAW_CARD_TYPE_FULL,
		Currency:   "",
	}

	if err = InsertWithdrawalCardUseByTx(wcu, tx); err != nil {
		tx.Rollback()
		return err
	}
	//临时额度
	if wr := InitUserWithdrawalByTx(uid, tx); wr != nil {
		if ok, err := IncomeUserWithdrawalCasualQuotaByTx(uid, card.Quota, tx); !ok {
			tx.Rollback()
			return err
		}
	} else {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil

}

func convUserWithdrawCard(row map[string]string) *UserWithdrawCard {
	if len(row) == 0 {
		return nil
	}
	uwc := &UserWithdrawCard{
		Id:         utils.Str2Int64(row["id"]),
		Password:   row["password"],
		TradeNo:    row["trade_no"],
		OwnerUid:   utils.Str2Int64(row["owner_uid"]),
		Quota:      utils.Str2Int64(row["quota"]),
		CreateTime: utils.Str2Int64(row["create_time"]),
		ExpireTime: utils.Str2Int64(row["expire_time"]),
		Cost:       utils.Str2Int64(row["cost"]),
		GetTime:    utils.Str2Int64(row["get_time"]),
		UseTime:    utils.Str2Int64(row["use_time"]),
		Status:     utils.Str2Int(row["status"]),
	}

	return uwc
}

func lvtc2BsvInMysql(uid int64, tx *sql.Tx) (int64, error) {

	_, err := tx.Exec("update user_asset_lvtc set `locked` = 0 where uid = ?  ", uid)

	if err != nil {
		logger.Error("modify user_asset_lvtc error", err.Error())
		return 0, err
	}

	_, err = tx.Exec("delete from user_asset_lock where uid = ? and currency = ?  ", uid,CURRENCY_LVTC)

	if err != nil {
		logger.Error("delete user_asset_lock lvtc error", err.Error())
		return 0, err
	}

	_, err = tx.Exec("delete from user_hashrate where uid = ? and type = 1  ", uid)

	if err != nil {
		logger.Error("delete user_hashrate lvtc error", err.Error())
		return 0, err
	}

	balance := int64(0)
	row := tx.QueryRow("select balance from user_asset_lvtc where uid = ?", uid)
	err = row.Scan(&balance)
	if err != nil {
		logger.Error("query lvtc balance error", err.Error())
		return 0, err
	}

	return balance,  nil

}

func lvt2LvtcInMysql(uid int64, tx *sql.Tx) (int64, int64, error) {

	ts := utils.GetTimestamp13()
	//资产锁定
	_, err := tx.Exec("select * from user_asset where uid = ? for update", uid)
	if err != nil {
		logger.Error("lock table error", err.Error())
		return 0, 0, err
	}
	_, err = tx.Exec("select * from user_asset_lvtc where uid = ? for update", uid)
	if err != nil {
		logger.Error("lock table error", err.Error())
		return 0, 0, err
	}
	_, err = tx.Exec("select * from user_asset_lock where uid = ? and currency = ? for update", uid, CURRENCY_LVT)
	if err != nil {
		logger.Error("lock table error", err.Error())
		return 0, 0, err
	}
	balance := int64(0)
	incomeAsset := int64(0)
	row := tx.QueryRow("select balance,income from user_asset where uid = ?", uid)
	err = row.Scan(&balance, &incomeAsset)
	if err != nil {
		logger.Error("query balance error", err.Error())
		return 0, 0, err
	}
	//余额不足的不处理
	if balance <= 0 {
		return 0, 0, nil
	}
	//获取转换汇率
	lvtcHashrateScale := int64(config.GetConfig().LvtcHashrateScale)

	income := 1
	userScore, _ := rpc.GetUserField(uid, microuser.UserField_CREDIT_SCORE)
	userLevel, _ := rpc.GetUserField(uid, microuser.UserField_LEVEL)
	if utils.Str2Int(userScore) >= DEF_SCORE && utils.Str2Int(userLevel) >= DEF_LEVEL {
		income = 0
	}
	//锁仓转换
	_, err = tx.Exec("update user_asset_lock set `value` = `value` / ?,currency = ?,income = ?  where uid = ? and currency = ? ", lvtcHashrateScale, CURRENCY_LVTC, income, uid, CURRENCY_LVT)

	if err != nil {
		logger.Error("modify user_asset_lock error", err.Error())
		return 0, 0, err
	}

	//获取转换后的锁仓锁定总额
	lvtcLockCount := int64(0)
	row = tx.QueryRow("select if(sum(value) is null,0,sum(value)) as c from user_asset_lock where uid = ? and currency = ?", uid, CURRENCY_LVTC)
	err = row.Scan(&lvtcLockCount)
	if err != nil {
		logger.Error("query asset lvtc lock count error", err.Error())
		return 0, 0, err
	}

	//修改锁仓
	//此处不对余额进行处理，交由上层统一走转账流程
	incomeAsset = incomeAsset / lvtcHashrateScale
	var updateLvtcSql string
	var updateLvtcParam []interface{}
	if incomeAsset > 0 {
		updateLvtcSql = "update user_asset_lvtc set locked = ?,income = income + ?,lastmodify = ? where uid = ?"
		updateLvtcParam = []interface{}{lvtcLockCount, incomeAsset, ts, uid}
	} else {
		updateLvtcSql = "update user_asset_lvtc set locked = ?,lastmodify = ? where uid = ?"
		updateLvtcParam = []interface{}{lvtcLockCount, ts, uid}
	}

	_, err = tx.Exec(updateLvtcSql, updateLvtcParam...)
	if err != nil {
		logger.Error("modify locked error", err.Error())
		return 0, 0, err
	}

	//修改锁仓
	//此处不对余额进行处理，交由上层统一走转账流程
	_, err = tx.Exec("update user_asset set locked = 0, income = 0,lastmodify = ? where uid = ?", ts, uid)
	if err != nil {
		logger.Error("modify locked error", err.Error())
		return 0, 0, err
	}
	//计算转换后的资产
	lvtcBalance := balance / lvtcHashrateScale

	return balance, lvtcBalance, nil

}

func lvt2LvtcDelayInMysql(uid int64, tx *sql.Tx) (int64, int64, error) {

	ts := utils.GetTimestamp13()
	//资产锁定 并查询余额
	balance := int64(0)
	row := tx.QueryRow("select balance from user_asset where uid = ? for update", uid)
	err := row.Scan(&balance)
	if err != nil {
		logger.Error("query balance error", err.Error())
		return 0, 0, err
	}
	//余额不足的不处理
	if balance <= 0 {
		return 0, 0, nil
	}

	_, err = tx.Exec("select * from user_asset_lvtc where uid = ? for update", uid)
	if err != nil {
		logger.Error("lock table error", err.Error())
		return 0, 0, err
	}
	//锁仓解除(lvt)
	_, err = tx.Exec("delete from user_asset_lock where uid = ? and currency = ? ", uid, CURRENCY_LVT)
	if err != nil {
		logger.Error("modify user_asset_lock error", err.Error())
		return 0, 0, err
	}

	_, err = tx.Exec("update user_asset set locked = 0,income = 0,lastmodify = ? where uid = ?", ts, uid)
	if err != nil {
		logger.Error("modify locked error", err.Error())
		return 0, 0, err
	}

	//前19个 自动锁仓

	begin := utils.GetTimestamp13()
	//分20期 取小数点后六位
	lockValue := balance / 20 / 100 * 100
	lockValueStr := utils.LVTintToFloatStr(lockValue)
	lvtScale := config.GetConfig().LvtcHashrateScale

	income := 1
	userScore, _ := rpc.GetUserField(uid, microuser.UserField_CREDIT_SCORE)
	userLevel, _ := rpc.GetUserField(uid, microuser.UserField_LEVEL)
	if utils.Str2Int(userScore) >= DEF_SCORE && utils.Str2Int(userLevel) >= DEF_LEVEL {
		income = 0
	}
	for i := 0; i < 19; i++ {
		month := i + 5
		//计算结束时间
		end := begin + (int64(month) * constants.ASSET_LOCK_MONTH_TIMESTAMP)

		assetLock := &AssetLockLvtc{
			Uid:         uid,
			Value:       lockValueStr,
			ValueInt:    lockValue,
			Month:       month,
			Hashrate:    utils.GetLockHashrate(lvtScale, month, lockValueStr),
			Begin:       begin,
			End:         end,
			Currency:    CURRENCY_LVTC,
			AllowUnlock: constants.ASSET_LOCK_UNLOCK_TYPE_ALLOW,
			Income:      income,
		}
		if err := CreateAssetLockConv(assetLock, tx); err != nil {
			logger.Error("Create Asset Lock error", err.Error())
			return 0, 0, err
		}
	}
	//最后一个锁仓的操作
	lastLockValue := lockValue + (balance - (lockValue * 20))
	lastLockValueStr := utils.LVTintToFloatStr(lastLockValue)
	month := 24
	//计算结束时间
	end := begin + (int64(month) * constants.ASSET_LOCK_MONTH_TIMESTAMP)

	assetLock := &AssetLockLvtc{
		Uid:         uid,
		Value:       lastLockValueStr,
		ValueInt:    lastLockValue,
		Month:       month,
		Hashrate:    utils.GetLockHashrate(lvtScale, month, lastLockValueStr),
		Begin:       begin,
		End:         end,
		Currency:    CURRENCY_LVTC,
		AllowUnlock: constants.ASSET_LOCK_UNLOCK_TYPE_ALLOW,
		Income:      income,
	}
	if err := CreateAssetLockConv(assetLock, tx); err != nil {
		logger.Error("Create Asset Lock error", err.Error())
		return 0, 0, err
	}

	//获取转换后的锁仓锁定总额
	lvtcLockCount := int64(0)
	row = tx.QueryRow("select if(sum(value) is null,0,sum(value)) as c from user_asset_lock where uid = ? and currency = ?", uid, CURRENCY_LVTC)
	err = row.Scan(&lvtcLockCount)
	if err != nil {
		logger.Error("query asset lvtc lock count error", err.Error())
		return 0, 0, err
	}

	//此处不对余额进行处理，交由上层统一走转账流程

	_, err = tx.Exec("update user_asset_lvtc set locked = ?,lastmodify = ? where uid = ?", lvtcLockCount, ts, uid)
	if err != nil {
		logger.Error("modify locked error", err.Error())
		return 0, 0, err
	}

	return balance, balance, nil

}

func CreateAssetLockConv(assetLock *AssetLockLvtc, tx *sql.Tx) error {
	//锁定记录
	tx.Exec("select * from user_asset_lvtc where uid = ? for update", assetLock.Uid)

	//修改资产数据
	//锁仓算力大于500时 给500
	updSql := `update
					user_asset_lvtc
			   set
			   		locked = locked + ?,
			   		lastmodify = ?
			   where
			   		uid = ?`
	updParams := []interface{}{
		assetLock.ValueInt,
		assetLock.Begin,
		assetLock.Uid,
	}
	info1, err1 := tx.Exec(updSql, updParams...)
	if err1 != nil {
		logger.Error("sql error ", err1.Error())
		return err1
	}
	//update 以后校验修改记录条数，如果为0 说明初始化部分出现问题，返回错误
	rsa, _ := info1.RowsAffected()
	if rsa == 0 {
		logger.Error("update user balance error RowsAffected ", rsa, " can not find user  ", assetLock.Uid, "")
		return sql.ErrNoRows
	}

	sql := "insert into user_asset_lock (uid,value,month,hashrate,begin,end,currency,allow_unlock,income) values (?,?,?,?,?,?,?,?,?)"
	params := []interface{}{
		assetLock.Uid,
		assetLock.ValueInt,
		assetLock.Month,
		assetLock.Hashrate,
		assetLock.Begin,
		assetLock.End,
		assetLock.Currency,
		assetLock.AllowUnlock,
		assetLock.Income,
	}
	res, err := tx.Exec(sql, params...)
	if err != nil {
		logger.Error("create asset lock error", err.Error())
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		logger.Error("get last insert id error", err.Error())
		return err
	}

	assetLock.Id = id
	assetLock.IdStr = utils.Int642Str(id)

	if ok, _ := updLockAssetHashRate(assetLock.Uid, tx); !ok {
		return errors.New("system error : update lock asset hashrate")
	}
	return nil
}

func ExtractIncomeLvtc(uid, income int64) bool {
	tx, err := gDBAsset.Begin()
	if err != nil {
		logger.Error("db pool begin error ", err.Error())
		return false
	}

	var ok bool

	if ok = ExtractIncomeLvtcByTx(uid, income, tx); ok {
		tx.Commit()
	} else {
		tx.Rollback()
	}
	return ok
}

func ExtractIncomeEth(uid, income int64) bool {
	tx, err := gDBAsset.Begin()
	if err != nil {
		logger.Error("db pool begin error ", err.Error())
		return false
	}

	var ok bool

	if ok = ExtractIncomeEthByTx(uid, income, tx); ok {
		tx.Commit()
	} else {
		tx.Rollback()
	}
	return ok
}

func ExtractIncomeLvtcByTx(uid, income int64, tx *sql.Tx) bool {
	tx.Exec("select * from user_asset_lvtc where uid = ? for update", uid)

	ts := utils.GetTimestamp13()

	//扣除提取income
	info, err := tx.Exec("update user_asset_lvtc set income = income - ?,lastmodify = ? where income >= ? and uid = ?", income, ts, income, uid)
	if err != nil {
		logger.Error("sql error ", err.Error())
		return false
	}
	//update 以后校验修改记录条数，如果为0 说明初始化部分出现问题，返回错误
	rsa, _ := info.RowsAffected()
	if rsa == 0 {
		logger.Error("update user income error RowsAffected ", rsa, " can not find user  ", uid, "")
		return false
	}

	return true
}

func ExtractIncomeEthByTx(uid, income int64, tx *sql.Tx) bool {
	tx.Exec("select * from user_asset_eth where uid = ? for update", uid)

	ts := utils.GetTimestamp13()

	//扣除提取income
	info, err := tx.Exec("update user_asset_eth set income = income - ?,lastmodify = ? where income >= ? and uid = ?", income, ts, income, uid)
	if err != nil {
		logger.Error("sql error ", err.Error())
		return false
	}
	//update 以后校验修改记录条数，如果为0 说明初始化部分出现问题，返回错误
	rsa, _ := info.RowsAffected()
	if rsa == 0 {
		logger.Error("update user income error RowsAffected ", rsa, " can not find user  ", uid, "")
		return false
	}

	return true
}

func GetMinerDays(uid int64) []int {
	sql := `
		select days from user_reward_lvtc where uid = ?
		union all
		select days from user_reward where uid = ?

	`
	rows := gDBAsset.Query(sql, uid, uid)
	if len(rows) > 0 {
		r := make([]int, 0)
		for _, v := range rows {
			r = append(r, utils.Str2Int(v["days"]))
		}
		return r
	}
	return nil
}

func MoveMinerDays(uid int64) int {
	dayss := GetMinerDays(uid)
	if dayss != nil || len(dayss) > 0 {
		for _, v := range dayss {
			if v > 0 {
				return v + 1
			}
		}
	}
	return 1
}

func QueryHashRateDetailByUid(uid int64) []map[string]string {
	sql := `select t.type, if(sum(t.h) is null,0,sum(t.h)) as sh from (
				select uh1.type, max(uh1.hashrate) as h from user_hashrate as uh1 where uh1.uid = ? and uh1.end = 0 group by uh1.type
				union all
				select uh2.type, uh2.hashrate as h from user_hashrate as uh2 where uh2.uid = ? and uh2.end >= ?
			) as t
			group by t.type
			`

	params := []interface{}{
		uid,
		uid,
		utils.GetTimestamp13(),
	}

	rows := gDBAsset.Query(sql, params...)

	return rows
}

func checkAssetBalanceIsSufficient(uid, amount, fee int64, currency, feeCurrency string) bool {
	assetTableName := ""
	switch strings.ToUpper(currency) {
	case CURRENCY_BTC:
		assetTableName = "user_asset_btc"
	case CURRENCY_ETH:
		assetTableName = "user_asset_eth"
	case CURRENCY_EOS:
		assetTableName = "user_asset_eos"
	case CURRENCY_LVTC:
		assetTableName = "user_asset_lvtc"
	}
	sql := fmt.Sprintf("select balance, locked, income from %s where uid = ?", assetTableName)
	row, err := gDBAsset.QueryRow(sql, uid)
	if err != nil {
		logger.Error("query asset error, uid:", uid, "error:", err)
	}
	if strings.EqualFold(currency, feeCurrency) {
		return utils.Str2Int64(row["balance"])-utils.Str2Int64(row["locked"])-utils.Str2Int64(row["income"]) > (amount + fee)
	} else {
		switch strings.ToUpper(feeCurrency) {
		case CURRENCY_BTC:
			assetTableName = "user_asset_btc"
		case CURRENCY_ETH:
			assetTableName = "user_asset_eth"
		case CURRENCY_EOS:
			assetTableName = "user_asset_eos"
		case CURRENCY_LVTC:
			assetTableName = "user_asset_lvtc"
		}
		sql = fmt.Sprintf("select balance, locked, income from %s where uid = ?", assetTableName)
		rowFee, err := gDBAsset.QueryRow(sql, uid)
		if err != nil {
			logger.Error("query asset error, uid:", uid, "error:", err)
		}
		return (utils.Str2Int64(row["balance"])-utils.Str2Int64(row["locked"])-utils.Str2Int64(row["income"]) > amount) &&
			(utils.Str2Int64(rowFee["balance"])-utils.Str2Int64(rowFee["locked"])-utils.Str2Int64(rowFee["income"]) > fee)
	}
}

func calculationFeeAndCheckQuotaForTransfer(uid int64, amount, currency, feeCurrency string, currencyDecimal int) (float64, constants.Error) {
	if utils.Str2Float64(amount) <= 0 {
		return float64(0), constants.RC_PARAM_ERR
	}
	transferQuota := getTransferQuota(currency, feeCurrency)
	if transferQuota == nil {
		return float64(0), constants.RC_PARAM_ERR
	}
	if transferQuota.SingleAmountMin > 0 && transferQuota.SingleAmountMin > utils.Str2Float64(amount) {
		return float64(0), constants.RC_TRANS_AMOUNT_EXCEEDING_LIMIT
	}

	if transferQuota.DailyAmountMax > 0 {
		var totalAmount int64
		if strings.EqualFold(currency, CURRENCY_LVT) {
			totalAmount = GetCurrentDayLVTTransferAmount(uid)
		}
		if strings.EqualFold(currency, CURRENCY_LVTC) {
			totalAmount = GetCurrentDayLVTCTransferAmount(uid)
		}
		if strings.EqualFold(currency, CURRENCY_ETH) || strings.EqualFold(currency, CURRENCY_EOS) || strings.EqualFold(currency, CURRENCY_BTC) {
			historyTableName := ""
			switch strings.ToUpper(currency) {
			case CURRENCY_BTC:
				historyTableName = "tx_history_btc"
			case CURRENCY_ETH:
				historyTableName = "tx_history_eth"
			case CURRENCY_EOS:
				historyTableName = "tx_history_eos"
			}
			sql := fmt.Sprintf("select sum(value) total_value from %s where `from` = ? and ts >= ?", historyTableName)
			row, err := gDBAsset.QueryRow(sql, uid, utils.GetTimestamp13ByTime(utils.GetDayStart(utils.GetTimestamp13())))
			if err != nil {
				logger.Error("query that day total transfer amount error, uid:", uid, ",error:", err.Error())
			}
			totalAmount = utils.Str2Int64(row["total_value"])
		}

		dailyAmount, _ := decimal.NewFromString(utils.Float642Str(transferQuota.DailyAmountMax))
		dailyAmountInt64 := dailyAmount.Mul(decimal.NewFromFloat(float64(currencyDecimal))).IntPart()
		amountBig, _ := decimal.NewFromString(amount)
		amountInt64 := amountBig.Mul(decimal.NewFromFloat(float64(currencyDecimal))).IntPart()

		logger.Info("daily quota out limit, transfer amount:", amountInt64, ",that day transfer total amount:", totalAmount, ",daily amount:", dailyAmountInt64)
		if dailyAmountInt64 < amountInt64+totalAmount {
			return float64(0), constants.RC_TRANS_AMOUNT_EXCEEDING_LIMIT
		}
	}
	feeOfTransfer := utils.Str2Float64(amount) * transferQuota.Fee.FeeRate
	if feeOfTransfer < transferQuota.Fee.FeeMin {
		feeOfTransfer = transferQuota.Fee.FeeMin
	}
	if feeOfTransfer > transferQuota.Fee.FeeMax {
		feeOfTransfer = transferQuota.Fee.FeeMax
	}

	if strings.ToUpper(currency) != strings.ToUpper(feeCurrency) {
		feeOfTransfer = ConversionCoinPrice(feeOfTransfer, strings.ToUpper(currency), strings.ToUpper(feeCurrency))
	}

	feeOfTransfer *= transferQuota.Fee.Discount

	if feeOfTransfer < 0 {
		logger.Error("transfer fee err, fee currency:", feeCurrency, "fee:", feeOfTransfer)
		return float64(0), constants.RC_TRANSFER_FEE_ERROR
	}

	return feeOfTransfer, constants.RC_OK
}

func getTransferQuota(currency, feeCurrency string) *TransferQuota {
	key := constants.TRANSFER_QUOTA_FEE_KEY + "_" + currency + "_" + feeCurrency
	quota := new(TransferQuota)
	results, _ := redis.String(rdsDo("GET", key))
	reload := false
	logger.Info("getTransferQuota:: key",key,"rds res",results)
	if len(results) > 0 {
		if err := utils.FromJson(results, quota); err != nil {
			logger.Error("get withdraw quota from redis error, error:", err.Error())
			reload = true
		}
	} else {
		reload = true
	}
	if reload {
		quota = GeTransferQuotaByCurrency(currency, feeCurrency)
		tomorrow, _ := time.ParseInLocation("2006-01-02", time.Now().Format("2006-01-02")+" 23:59:59", time.Local)
		expire := tomorrow.Unix()+1-utils.GetTimestamp10()
		jsonStr := utils.ToJSON(quota)
		rdsDo("SET", key, jsonStr, "EX",expire )
		logger.Info("getTransferQuota: key",key,"reload into rds json",jsonStr,"expire",expire)
	}
	return quota
}
