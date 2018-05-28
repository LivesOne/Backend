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
)

const (
	CONV_LVT          = 10000 * 10000
	DAY_QUOTA_TYPE    = 0
	CASUAL_QUOTA_TYPE = 1
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
			incomeCasual = assetLock.ValueInt/2
		case 12:
			incomeCasual = assetLock.ValueInt
		default:
			tx.Rollback()
			return false, constants.TRANS_ERR_SYS
		}

		if wr := InitUserWithdrawal(assetLock.Uid);wr != nil {
			if ok,_ := IncomeUserWithdrawalCasualQuota(assetLock.Uid,incomeCasual);!ok{
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
		incomeCasual = assetLock.ValueInt/2
	case 12:
		incomeCasual = assetLock.ValueInt
	default:
		tx.Rollback()
		return false, constants.TRANS_ERR_SYS
	}

	if wr := InitUserWithdrawal(assetLock.Uid);wr != nil {
		if ok,_ := IncomeUserWithdrawalCasualQuota(assetLock.Uid,incomeCasual);!ok{
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
		Casual:    utils.Str2Int64(al["month"]),
		DayExpend: utils.Str2Int64(al["day_expend"]),
	}
	return &alres
}

func CreateUserWithdrawalQuota(uid int64, day int64, month int64) (sql.Result, error) {
	sql := "insert ignore into user_withdrawal_quota(uid, `day`, `month`, casual, day_expend, last_expend, last_income) values(?, ?, ?, ?, ?, ?, ?) "
	return gDBAsset.Exec(sql, uid, day, month, 0, 0, utils.GetTimestamp13(), 0)

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

func ExpendUserWithdrawalQuota(uid int64, expendQuota int64, quotaType int) (bool, error) {
	if expendQuota <= 0 && quotaType > 0 {
		return false, errors.New("expend quota must greater than 0")
	}

	if quotaType != DAY_QUOTA_TYPE && quotaType != CASUAL_QUOTA_TYPE {

	}

	if quotaType == DAY_QUOTA_TYPE {
		sql := "update user_withdrawal_quota set day = day - ?,month = month - ?,day_expend = ?,last_expend = ? where uid = ? and day > ? and month > ?"
		result, err := gDBAsset.Exec(sql, expendQuota, expendQuota, utils.GetTimestamp13(), utils.GetTimestamp13(), uid, expendQuota, expendQuota)
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected > 0 {
			return true, nil
		} else {
			return false, err
		}
	}
	if quotaType != CASUAL_QUOTA_TYPE {
		sql := "update user_withdrawal_quota set casual = casual - ?,last_expend = ? where uid = ? and casual > ?"
		result, err := gDBAsset.Exec(sql, expendQuota, utils.GetTimestamp13(), uid, expendQuota)
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected > 0 {
			return true, nil
		} else {
			return false, err
		}
	}
	return false, errors.New("record not exist")

}

func IncomeUserWithdrawalCasualQuota(uid int64, incomeCasual int64) (bool, error) {
	if incomeCasual > 0 {
		sql := "update user_withdrawal_quota set casual = casual + ?,last_expend = ? where uid = ?"
		result, err := gDBAsset.Exec(sql, incomeCasual, utils.GetTimestamp13(), uid)
		if err != nil {
			logger.Error("exec sql error",sql)
			return false, err
		}
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected > 0 {
			return true, nil
		} else {
			return false, sqlBase.ErrNoRows
		}
	}
	return false, errors.New("income casual quota must greater then 0")
}


func InitUserWithdrawal(uid int64)*UserWithdrawalQuota{
	level := GetTransUserLevel(uid)
	limitConfig := config.GetLimitByLevel(level)
	_,err := CreateUserWithdrawalQuota(uid, utils.FloatStrToLVTint(utils.Int642Str(limitConfig.DailyWithdrawalQuota())), utils.FloatStrToLVTint(utils.Int642Str(limitConfig.MonthlyWithdrawalQuota())))
	if err != nil {
		logger.Error("insert user withdrawal quota error for user:" , uid)
		return nil
	}
	return &UserWithdrawalQuota{
		Day:       utils.FloatStrToLVTint(utils.Int642Str(limitConfig.DailyWithdrawalQuota())),
		Month:     utils.FloatStrToLVTint(utils.Int642Str(limitConfig.MonthlyWithdrawalQuota())),
		Casual:    0,
		DayExpend: 0,
	}
}

