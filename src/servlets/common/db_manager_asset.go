package common

import (
	"database/sql"
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"servlets/constants"
	"utils"
	"utils/config"
	"utils/db_factory"
	"utils/logger"
	sqlBase "database/sql"
	"time"
	"encoding/json"
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
		resReward.Days = utils.Str2Int(row["days"])

	}
	return resReward, err

}

func QueryBalance(uid int64) (int64, int64, error) {
	row, err := gDBAsset.QueryRow("select balance,locked from user_asset where uid = ?", uid)
	if err != nil {
		logger.Error("query db error ", err.Error())
	}

	if row != nil {
		return utils.Str2Int64(row["balance"]), utils.Str2Int64(row["locked"]), nil
	}
	return 0, 0, err
}

func QueryBalanceEth(uid int64) (int64, int64, error) {
	row, err := gDBAsset.QueryRow("select balance,locked from user_asset_eth where uid = ?", uid)
	if err != nil {
		logger.Error("query db error ", err.Error())
	}

	if row != nil {
		return utils.Str2Int64(row["balance"]), utils.Str2Int64(row["locked"]), nil
	}
	return 0, 0, err
}

func TransAccountLvt(txid, from, to, value int64) (bool, int) {
	//检测资产初始化情况
	//from 的资产如果没有初始化，初始化并返回false--》 上层检测到false会返回余额不足
	f, c := CheckAndInitAsset(from)
	if !f {
		return f, c
	}

	tx, err := gDBAsset.Begin()
	if err != nil {
		logger.Error("db pool begin error ", err.Error())
		return false, constants.TRANS_ERR_SYS
	}

	var (
		ok bool
		e  int
	)

	if ok, e = TransAccountLvtByTx(txid, from, to, value, tx); ok {
		tx.Commit()
	} else {
		tx.Rollback()
	}
	return ok, e
}

func TransAccountLvtByTx(txid, from, to, value int64, tx *sql.Tx) (bool, int) {
	tx.Exec("select * from user_asset where uid in (?,?) for update", from, to)

	//资产冻结状态校验，如果status是0 返回true 继续执行，status ！= 0 账户冻结，返回错误
	if !CheckAssetLimeted(from, tx) {
		tx.Rollback()
		return false, constants.TRANS_ERR_ASSET_LIMITED
	}
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

	_, err = InsertAssetEth(uid)
	if err != nil {
		logger.Error("init asset eth error ", err.Error())
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
func InsertAssetEth(uid int64) (sql.Result, error) {
	sql := "insert ignore into user_asset_eth (uid,balance,lastmodify) values (?,?,?) "
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

func ckeckBalance(uid int64, value int64, tx *sql.Tx) bool {
	var balance int64
	var locked int64
	row := tx.QueryRow("select balance,locked from user_asset where uid  = ?", uid)
	row.Scan(&balance, &locked)
	return balance > 0 && (balance-locked) >= value
}

func ckeckEthBalance(uid int64, value int64, tx *sql.Tx) bool {
	var balance int64
	var locked int64
	row := tx.QueryRow("select balance,locked from user_asset_eth where uid  = ?", uid)
	row.Scan(&balance, &locked)
	return balance > 0 && (balance-locked) >= value
}

func CreateAssetLock(assetLock *AssetLock) (bool, int) {

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

	//修改资产数据
	//锁仓算力大于500时 给500
	updSql := `update
					user_asset
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

	sql := "insert into user_asset_lock (uid,value,month,hashrate,begin,end,type) values (?,?,?,?,?,?,?)"
	params := []interface{}{
		assetLock.Uid,
		assetLock.ValueInt,
		assetLock.Month,
		assetLock.Hashrate,
		assetLock.Begin,
		assetLock.End,
		assetLock.Type,
	}
	res, err := tx.Exec(sql, params...)
	if err != nil {
		logger.Error("create asset lock error", err.Error())
		tx.Rollback()
		return false, constants.TRANS_ERR_SYS
	}
	id, err := res.LastInsertId()
	if err != nil {
		logger.Error("get last insert id error", err.Error())
		tx.Rollback()
		return false, constants.TRANS_ERR_SYS
	}

	if ok, code := updLockAssetHashRate(assetLock.Uid, tx); !ok {
		tx.Rollback()
		return ok, code
	}
	if assetLock.Type == ASSET_LOCK_TYPE_DRAW {

		incomeCasual := int64(0)
		switch assetLock.Month {
		case 6:
			incomeCasual = assetLock.ValueInt / 2
		case 12:
			incomeCasual = assetLock.ValueInt
		default:
			tx.Rollback()
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

	}
	tx.Commit()
	assetLock.Id = id
	assetLock.IdStr = utils.Int642Str(id)
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

func QueryAssetLockList(uid int64) []*AssetLock {
	res := gDBAsset.Query("select * from user_asset_lock where uid = ? order by id desc", uid)
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

func convAssetLockList(list []map[string]string) []*AssetLock {
	listRes := make([]*AssetLock, 0)
	for _, v := range list {
		listRes = append(listRes, convAssetLock(v))
	}
	return listRes
}

func execRemoveAssetLock(txid int64, assetLock *AssetLock, penaltyMoney int64, tx *sql.Tx) (bool, int) {
	//锁定记录
	ts := utils.TXIDToTimeStamp13(txid)
	to := config.GetConfig().PenaltyMoneyAccountUid
	tx.Exec("select * from user_asset where uid in (?,?) for update", assetLock.Uid, to)

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

func RemoveAssetLock(txid int64, assetLock *AssetLock, penaltyMoney int64) (bool, int) {
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

func QuerySumLockAsset(uid int64, month int) int64 {
	row, err := gDBAsset.QueryRow("select if(sum(value) is null,0,sum(value)) as value from user_asset_lock where uid = ? and month >= ?", uid, month)
	if err != nil {
		logger.Error("query user asset lock error", err.Error())
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
	row, err := gDBAsset.QueryRow("SELECT uid,`day`,`month`,casual,day_expend FROM user_withdrawal_quota where uid = ?", uid)

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
	}
	return &alres
}

func CreateUserWithdrawalQuota(uid int64, day int64, month int64) (sql.Result, error) {
	return CreateUserWithdrawalQuotaByTx(uid, day, month, nil)
}
func CreateUserWithdrawalQuotaByTx(uid int64, day int64, month int64, tx *sql.Tx) (sql.Result, error) {
	if tx == nil {
		tx, _ = gDBAsset.Begin()
		defer tx.Commit()
	}
	sql := "insert ignore into user_withdrawal_quota(uid, `day`, `month`, casual, day_expend, last_expend, last_income) values(?, ?, ?, ?, ?, ?, ?) "
	return tx.Exec(sql, uid, day, month, 0, 0, utils.GetTimestamp13(), 0)
}

func ResetDayQuota(uid int64, dayQuota int64) bool {
	sql := "update user_withdrawal_quota set `day` = ? where uid = ?"
	result, err := gDBAsset.Exec(sql, dayQuota, uid)
	if err != nil {
		logger.Error("重置月额度错误" + err.Error())
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		return true
	} else {
		return false
	}
}

func ResetMonthQuota(uid int64, monthQuota int64) bool {
	sql := "update user_withdrawal_quota set `month` = ? where uid = ?"
	result, err := gDBAsset.Exec(sql, monthQuota, uid)
	if err != nil {
		logger.Error("重置月额度错误" + err.Error())
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		return true
	} else {
		return false
	}
}

func ExpendUserWithdrawalQuota(uid int64, expendQuota int64, quotaType int, tx *sql.Tx) (bool, error) {
	logger.Debug("扣除提币额度开始")
	if expendQuota <= 0 {
		logger.Debug("扣除提币额度错误")
		return false, errors.New("expend quota must greater than 0")
	}

	if quotaType != DAY_QUOTA_TYPE && quotaType != CASUAL_QUOTA_TYPE {
		logger.Debug("提币额度类型错误")
		return false, errors.New("expend quota type error")
	}

	if quotaType == DAY_QUOTA_TYPE {
		sql := "update user_withdrawal_quota set day = day - ?,month = month - ?,day_expend = ?,last_expend = ? where uid = ? and day > ? and month > ?"
		result, err := gDBAsset.Exec(sql, expendQuota, expendQuota, utils.GetTimestamp13(), utils.GetTimestamp13(), uid, expendQuota, expendQuota)
		if err != nil {
			logger.Error("扣除日额度错误", err.Error())
			return false, err
		}
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected > 0 {
			return true, nil
		} else {
			logger.Debug("扣除当日额度错误", err.Error())
			return false, err
		}
	}
	if quotaType == CASUAL_QUOTA_TYPE {
		sql := "update user_withdrawal_quota set casual = casual - ?,last_expend = ? where uid = ? and casual > ?"
		result, err := gDBAsset.Exec(sql, expendQuota, utils.GetTimestamp13(), uid, expendQuota)
		if err != nil {
			logger.Error("扣除临时额度错误", err.Error())
			return false, err
		}
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected > 0 {
			return true, nil
		} else {
			logger.Debug("扣除临时额度错误", err.Error())
			return false, err
		}
	}
	logger.Debug("扣除额度未知错误")
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
	level := GetTransUserLevel(uid)
	limitConfig := config.GetLimitByLevel(level)
	day, month := limitConfig.DailyWithdrawalQuota()*utils.CONV_LVT,
		limitConfig.MonthlyWithdrawalQuota()*utils.CONV_LVT
	_, err := CreateUserWithdrawalQuota(uid, day, month)
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

func InitUserWithdrawalByTx(uid int64, tx *sql.Tx) *UserWithdrawalQuota {
	level := GetTransUserLevel(uid)
	limitConfig := config.GetLimitByLevel(level)
	day, month := limitConfig.DailyWithdrawalQuota()*utils.CONV_LVT,
		limitConfig.MonthlyWithdrawalQuota()*utils.CONV_LVT
	_, err := CreateUserWithdrawalQuotaByTx(uid, day, month, tx)
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

func QueryWithdrawValueOfCurMonth(uid int64) int64 {
	sql := "select sum(value) from user_withdrawal_request where uid = ? and create_time > ?"

	tm := time.Unix(time.Now().Unix(), 0)
	tm2, _ := time.Parse("2006-01", tm.Format("2006-01"))
	result, _ := gDBAsset.QueryRow(sql, uid, utils.GetTimestamp13ByTime(tm2))
	return utils.Str2Int64(result["value"])
}

func Withdraw(uid int64, amount int64, address string, quotaType int) (string, constants.Error) {
	tx, _ := gDBAsset.Begin()
	logger.Debug("开始提币申请事务")
	row := tx.QueryRow("select count(1) from user_withdrawal_request where uid = ? and status in (?, ?, ?)", uid, constants.USER_WITHDRAWAL_REQUEST_WAIT_SEND, constants.USER_WITHDRAWAL_REQUEST_SEND, constants.USER_WITHDRAWAL_REQUEST_UNKNOWN)
	processingCount := int64(-1)
	errQuery := row.Scan(&processingCount)
	if errQuery != nil {
		logger.Error("query day_expend from user_withdrawal_quota error ")
		tx.Rollback()
		return "", constants.RC_SYSTEM_ERR
	}

	if processingCount > 0 {
		logger.Debug("有处理中的提币申请")
		tx.Rollback()
		return "", constants.RC_HAS_UNFINISHED_WITHDRAWAL_TASK
	}

	tx.Exec("select * from user_withdrawal_request where uid = ? for update", uid)

	tradeNo := GenerateTradeNo(quotaType, quotaType) //TODO 修改

	flag, err := ExpendUserWithdrawalQuota(uid, amount, quotaType, tx)
	if err != nil {
		logger.Debug("扣除提币额度错误")
		return "", constants.RC_PARAM_ERR
	}
	logger.Debug("扣除提币额度完成")
	if flag {
		timestamp := utils.GetTimestamp13()
		txid_lvt := GenerateTxID()
		toLvt := config.GetWithdrawalConfig().LvtAcceptAccount
		logger.Debug("扣除LVT资产开始")
		transLvtResult, e := TransAccountLvtByTx(txid_lvt, uid, toLvt, amount, tx)
		if !transLvtResult {
			tx.Rollback()
			switch e {
			case constants.TRANS_ERR_INSUFFICIENT_BALANCE:
				return "", constants.RC_INSUFFICIENT_BALANCE
			case constants.TRANS_ERR_SYS:
				return "", constants.RC_TRANS_IN_PROGRESS
			case constants.TRANS_ERR_ASSET_LIMITED:
				return "", constants.RC_ACCOUNT_ACCESS_LIMITED
			default:
				return "", constants.RC_SYSTEM_ERR
			}
		}
		_, err3 := tx.Exec("insert into tx_history_lvt_tmp (txid, type, trade_no, `from`, `to`, value, ts) VALUES (?, ?, ?, ?, ?, ?, ?)", txid_lvt, constants.TX_TYPE_WITHDRAW_LVT, tradeNo, uid, toLvt, amount, timestamp)
		if err3 != nil {
			logger.Error("query error ", err.Error())
		}
		logger.Debug("扣除LVT资产完成")
		logger.Debug("扣除ETH资产开始")
		toEth := config.GetWithdrawalConfig().EthAcceptAccount
		txid_eth, e := EthTransCommit(uid, toEth, amount, tradeNo, constants.TX_TYPE_WITHDRAW_ETH_FEE, tx)
		if txid_eth <= 0 {
			tx.Rollback()
			switch e {
			case constants.TRANS_ERR_INSUFFICIENT_BALANCE:
				return "", constants.RC_INSUFFICIENT_BALANCE
			case constants.TRANS_ERR_SYS:
				return "", constants.RC_TRANS_IN_PROGRESS
			case constants.TRANS_ERR_ASSET_LIMITED:
				return "", constants.RC_ACCOUNT_ACCESS_LIMITED
			default:
				return "", constants.RC_SYSTEM_ERR
			}
		}
		logger.Debug("扣除ETH资产完成")
		logger.Debug("插入提币申请开始")

		sql := "insert into user_withdrawal_request (trade_no, uid, value, address, txid_lvt, txid_eth, create_time, update_time, status) values(?, ?, ?, ?, ?,?, ?, ?, ?)"
		_, err1 := tx.Exec(sql, tradeNo, uid, amount, address, txid_lvt, txid_eth, timestamp, timestamp, 0)
		if err1 != nil {
			logger.Error("add user_withdrawal_request error ", err.Error())
			flag = false
		}
		logger.Debug("插入提币申请完成")
		tx.Commit()

		//同步至mongo
		go func() {
			txh := &DTTXHistory{
				Id:      txid_lvt,
				TradeNo: tradeNo,
				Type:    constants.TX_TYPE_WITHDRAW_LVT,
				From:    uid,
				To:      toLvt,
				Value:   amount,
				Ts:      timestamp,
			}
			err = InsertCommited(txh)
			if err != nil {
				logger.Error("tx_history_lv_tmp insert mongo error ", err.Error())
				b, _ := json.Marshal(txh)
				rdsDo("rpush", constants.PUSH_TX_HISTORY_LVT_QUEUE_NAME, b)
			} else {
				DeleteTxhistoryLvtTmpByTxid(txid_lvt)
			}
		}()

		return tradeNo, constants.RC_OK
	} else {
		tx.Rollback()
		return "", constants.RC_HAS_UNFINISHED_WITHDRAWAL_TASK

	}

}

func DeleteTxhistoryLvtTmpByTxid(txid int64) {
	gDBAsset.Exec("delete from tx_history_lvt_tmp where txid= ? ", txid)
}

func QueryTxhistoryLvtTmpByTimie(ts int64) []*DTTXHistory {
	if gDBAsset != nil && gDBAsset.IsConn() {
		results := gDBAsset.Query("select * from tx_history_lvt_tmp where ts < ?", ts)
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
		Id:      utils.Str2Int64(al["txid"]),
		Type:    utils.Str2Int(al["type"]),
		TradeNo: al["trade_no"],
		From:    utils.Str2Int64(al["from"]),
		To:      utils.Str2Int64(al["to"]),
		Value:   utils.Str2Int64(al["value"]),
		Ts:      utils.Str2Int64(al["ts"]),
	}
	return &alres
}

func UpdateWithdrawResult(transId int64, result int) bool {
	tx, err := gDBAsset.Begin()
	if err != nil {
		logger.Error("db pool begin error ", err.Error())
		return false
	}

	var sql = ""
	if result == constants.USER_WITHDRAWAL_REQUEST_SUCCESS {
		//todo 更新提币申请状态
		sql = "update user_withdrawal_request set status = ? where id = ?"
		result, err := tx.Exec(sql, result, transId)
		if err != nil {
			logger.Error("update status in user_withdrawal_request error ", err.Error())
			tx.Rollback()
		}
		rowsAffected, err := result.RowsAffected()
		if rowsAffected != 1 {
			logger.Error("update status in user_withdrawal_request error ", err.Error())
			tx.Rollback()
		}
		tx.Commit()
	} else if result == constants.USER_WITHDRAWAL_REQUEST_FAIL {
		//RepayUserWithdrawalQuota(uid, repayQuota, quotaType, tx)
		//TODO 更新提币申请记录状态
		//TODO 新建反向交易，退还LVT
	}
	return true
}

/**
 * 退还提币额度
 */
func RepayUserWithdrawalQuota(uid int64, repayQuota int64, quotaType int, tx *sql.Tx) (bool, error) {
	var sql = ""
	if quotaType == CASUAL_QUOTA_TYPE {
		sql = "update user_withdrawal_quota set casual = casual + ?,last_expend = ? where uid = ?"
		result, err := tx.Exec(sql, repayQuota, utils.GetTimestamp13(), uid)
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected > 0 {
			return true, nil
		} else {
			return false, err
		}
	}

	result, err := tx.Exec(sql, repayQuota, utils.GetTimestamp13(), uid)
	if err != nil {
		logger.Error("exec sql error", sql)
		return false, err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		return true, nil
	} else {
		return false, sqlBase.ErrNoRows
	}
}

func QueryWithdrawalList(uid int64) []*UserWithdrawalRequest {
	sql := "select id,trade_no, uid, value, address, txid_lvt, txid_eth, create_time, update_time, case status when 0 then 1 else status end status from user_withdrawal_request where uid = ?"
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
		Uid:        utils.Str2Int64(al["uid"]),
		Address:    al["address"],
		Value:      utils.Str2Int64(al["value"]),
		TxidLvt:    utils.Str2Int64(al["txid_lvt"]),
		TxidEth:    utils.Str2Int64(al["txid_eth"]),
		CreateTime: utils.Str2Int64(al["create_time"]),
		UpdateTime: utils.Str2Int64(al["update_time"]),
		Status:     utils.Str2Int(al["status"]),
	}
	return &alres
}

func EthTransCommit(from, to, value int64, tradeNo string, tradeType int, tx *sql.Tx) (int64, int) {
	if tx == nil {
		tx, _ = gDBAsset.Begin()
		defer tx.Commit()
	}

	//检测资产初始化情况
	//from 的资产如果没有初始化，初始化并返回false--》 上层检测到false会返回余额不足
	logger.Debug("检测资产初始化情况开始")
	//f, c := CheckAndInitAsset(from)
	//if !f {
	//	return 0, c
	//}
	logger.Debug("检测资产初始化情况结束")
	tx.Exec("select * from user_asset_eth where uid in (?,?) for update", from, to)

	ts := utils.GetTimestamp13()

	//查询转出账户余额是否满足需要 使用新的校验方法，考虑到锁仓的问题
	logger.Debug("查询转出账户余额开始")
	if !ckeckEthBalance(from, value, tx) {
		return 0, constants.TRANS_ERR_INSUFFICIENT_BALANCE
	}
	logger.Debug("查询转出账户余额结束")
	logger.Debug("扣除转出方余额开始")
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
	logger.Debug("扣除转出方余额结束")
	logger.Debug("增加目标方余额开始")
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
	logger.Debug("增加目标方余额结束")
	txid := GenerateTxID()

	if txid == -1 {
		logger.Error("can not get txid")
		return 0, constants.TRANS_ERR_SYS
	}
	logger.Debug("插入tx_history_eth开始")
	info3, err3 := tx.Exec("insert into tx_history_eth (txid,type,trade_no,`from`,`to`,`value`,ts) values (?,?,?,?,?,?,?)",
		txid,
		tradeType,
		tradeNo,
		from,
		to,
		value,
		ts,
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
	logger.Debug("插入tx_history_eth完成")
	return txid, constants.TRANS_ERR_SUCC
}

func InsertTradePending(uid int64, tradeNo, bizContent string, value int64, tradeType int) error {
	_, err := gDBAsset.Exec("insert into trade_pending (trade_no,uid,type,biz_content,value,ts) values (?,?,?,?,?,?)", tradeNo, uid, tradeType, bizContent, value, utils.GetTimestamp13())
	return err
}

func GetTradePendingByTradeNo(tradeNo string, uid int64) (*TradePending, error) {
	sql := "select * from trade_pending where trade_no = ? and uid = ?"
	row, err := gDBAsset.QueryRow(sql, tradeNo, uid)
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
	tp.TradeNo = row["trade_no"]
	tp.BizContent = row["biz_content"]
	tp.Uid = utils.Str2Int64(row["uid"])
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
	_, err := tx.Exec("delete from trade_pending where trade_no = ? and uid = ?", tradeNo, uid)
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
	_, err := tx.Exec("insert into user_withdrawal_card_use (trade_no,uid,quota,cost,create_time,type) values (?,?,?,?,?,?)",
		wcu.TradeNo,
		wcu.Uid,
		wcu.Quota,
		wcu.Cost,
		wcu.CreateTime,
		wcu.Type,
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
	sql += " limit ?"
	params = append(params, max)
	rows := gDBAsset.Query(sql, params...)
	return rows
}

func GetUserWithdrawCardByUid(uid int64) []map[string]string {
	rows := gDBAsset.Query("select * from withdrawal_card where owner_uid = ?", uid)
	return rows
}

func GetUserWithdrawCardUseByUid(uid int64) []map[string]string {
	rows := gDBAsset.Query("select * from user_withdrawal_card_use where uid = ?", uid)
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
	tradeNo := GenerateTradeNo(1, 1)
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
