package common

import (
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"regexp"
	"utils"
	"utils/config"
	"utils/db_factory"
	"utils/logger"
	"servlets/constants"
	"database/sql"
)

//var gDbUser *sql.DB
var gDbUser *db_factory.DBPool

func UserDbInit() error {

	//dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?timeout=90s&charset=utf8",
	//	config.GetConfig().DBUser,
	//	config.GetConfig().DBUserPwd,
	//	config.GetConfig().DBHost,
	//	config.GetConfig().DBDatabase)

	//dsn := "lvt_serv:U9D$WHTo@(10.100.15.96:3306)/livesone_user?charset=utf8"
	//
	//var err error
	//gDbUser, err = sql.Open("mysql", dsn)
	//if err != nil {
	//	logger.Error(err)
	//	return err
	//}
	db_config_user := config.GetConfig().User
	facConfig_user := db_factory.Config{
		Host:        db_config_user.DBHost,
		UserName:    db_config_user.DBUser,
		Password:    db_config_user.DBUserPwd,
		Database:    db_config_user.DBDatabase,
		MaxConn:     db_config_user.MaxConn,
		MaxIdleConn: db_config_user.MaxConn,
	}
	gDbUser = db_factory.NewDataSource(facConfig_user)
	if gDbUser.IsConn() {
		logger.Debug("connection ", db_config_user.DBHost, db_config_user.DBDatabase, "database successful")
	} else {
		logger.Error(gDbUser.Err())
		return gDbUser.Err()
	}

	return nil
}

func ExistsUID(uid int64) bool {
	row, _ := gDbUser.QueryRow("select count(1) as c from account where uid = ? limit 1", uid)
	if row == nil {
		return false
	}
	return utils.Str2Int(row["c"]) > 0
}

func ExistsEmail(email string) bool {
	row, _ := gDbUser.QueryRow("select count(1) as c from account where email = ? limit 1", email)
	if row == nil {
		return false
	}
	return utils.Str2Int(row["c"]) > 0
}

func CheckUserLoginLimited(uid int64)bool{
	row, err := gDbUser.QueryRow("select status from account where uid = ? ", uid)
	if err != nil || row == nil {
		logger.Error("query err ", err.Error())
		return false
	}
	return utils.Str2Int(row["status"]) == constants.USER_LIMITED_DEF
}

func ExistsPhone(country int, phone string) bool {
	row, _ := gDbUser.QueryRow("select count(1) as c from account where country = ? and phone = ? limit 1", country, phone)
	if row == nil {
		return false
	}
	return utils.Str2Int(row["c"]) > 0
}

func CheckResetPhone(country int, phone string)bool{
	row, err := gDbUser.QueryRow("select uid from account where country = ? and phone = ? limit 1", country, phone)
	if err != nil || row == nil {
		return true
	}
	uid := utils.Str2Int64(row["uid"])
	flag,_ := CheckUserInLoginLimit(uid)
	return flag
}

func CheckResetEmail(email string) bool{
	row, err := gDbUser.QueryRow("select uid from account where email = ? limit 1", email)
	if err != nil || row == nil {
		return true
	}
	uid := utils.Str2Int64(row["uid"])
	flag,_ := CheckUserInLoginLimit(uid)
	return flag
}

func GetAssetByUid(uid int64) (int64, int) {
	row, _ := gDbUser.QueryRow("select uid,status from account where uid = ? limit 1", uid)
	if row == nil {
		return 0, 0
	}
	return utils.Str2Int64(row["uid"]), utils.Str2Int(row["status"])
}

func GetAssetByEmail(email string) (int64, int) {
	row, _ := gDbUser.QueryRow("select uid,status from account where email = ? limit 1", email)
	if row == nil {
		return 0, 0
	}
	return utils.Str2Int64(row["uid"]), utils.Str2Int(row["status"])
}

func GetAssetByPhone(country int, phone string) (int64, int) {
	row, _ := gDbUser.QueryRow("select uid,status from account where country = ? and phone = ? limit 1", country, phone)
	if row == nil {
		return 0, 0
	}
	return utils.Str2Int64(row["uid"]), utils.Str2Int(row["status"])
}

func InsertAccount(account *Account) (int64, error) {
	if !gDbUser.IsConn() {
		logger.Error("database not ready")
		return 0, errors.New("database not ready")
	}
	stmt, err := gDbUser.Prepare("INSERT account SET uid=?, login_password=?, " +
		"`language`=?, region=?, `from`=?, register_time=?, update_time=?, register_type=?")
	if err != nil {
		logger.Error(err)
		return 0, err
	}
	defer stmt.Close()
	ret, err := stmt.Exec(account.UID, account.LoginPassword,
		account.Language, account.Region, account.From, account.RegisterTime, account.UpdateTime, account.RegisterType)
	if err != nil {
		logger.Error(err)
		return 0, err
	}

	id, _ := ret.LastInsertId()

	return id, nil
}

func InsertAccountWithEmail(account *Account) (int64, error) {
	if !gDbUser.IsConn() {
		logger.Error("database not ready")
		return 0, errors.New("database not ready")
	}

	stmt, err := gDbUser.Prepare("INSERT account SET uid=?, email=?, login_password=?, " +
		"`language`=?, region=?, `from`=?, register_time=?, update_time=?, register_type=?")
	if err != nil {
		logger.Error(err)
		return 0, err
	}
	defer stmt.Close()
	ret, err := stmt.Exec(account.UID, account.Email, account.LoginPassword,
		account.Language, account.Region, account.From, account.RegisterTime, account.UpdateTime, account.RegisterType)
	if err != nil {
		logger.Error(err)
		return 0, err
	}

	id, _ := ret.LastInsertId()

	return id, nil
}

func InsertAccountWithPhone(account *Account) (int64, error) {
	if !gDbUser.IsConn() {
		logger.Error("database not ready")
		return 0, errors.New("database not ready")
	}

	stmt, err := gDbUser.Prepare("INSERT account SET uid=?, country=?, phone=?, login_password=?, " +
		"`language`=?, region=?, `from`=?, register_time=?, update_time=?, register_type=?")
	if err != nil {
		logger.Error(err)
		return 0, err
	}
	defer stmt.Close()
	ret, err := stmt.Exec(account.UID, account.Country, account.Phone, account.LoginPassword,
		account.Language, account.Region, account.From, account.RegisterTime, account.UpdateTime, account.RegisterType)
	if err != nil {
		logger.Error(err)
		return 0, err
	}

	id, _ := ret.LastInsertId()

	return id, nil
}

func GetAccountByUID(uid string) (*Account, error) {
	row, err := gDbUser.QueryRow("select * from account where uid = ?", uid)
	// logger.Info("GetAccountByUID:--------------------------", row, len(row), err)
	if err != nil {
		logger.Error(err)
		return nil,err
	}
	if len(row) < 1 {
		return nil, sql.ErrNoRows
	}
	return convRowMap2Account(row), err
}

func GetAccountByEmail(email string) (*Account, error) {
	row, err := gDbUser.QueryRow("select * from account where email = ?", email)
	// logger.Info("GetAccountByEmail:--------------------------", row, len(row), err)
	if err != nil {
		logger.Error(err)
	}
	if len(row) < 1 {
		return nil, errors.New("no record for:" + email)
	}
	return convRowMap2Account(row), err
}

func GetAccountByPhone(country int, phone string) (*Account, error) {
	row, err := gDbUser.QueryRow("select * from account where country = ? and phone = ? limit 1", country, phone)
	if err != nil {
		return nil, err
	}
	return convRowMap2Account(row), err
}

func GetAccountListByPhoneOnly(phone string) ([](*Account), error) {
	rows := gDbUser.Query("select * from account where phone = ? ", phone)
	// logger.Info("GetAccountListByPhoneOnly:--------------------------", rows, len(rows))
	if (rows == nil) || (len(rows) < 1) {
		return nil, errors.New("no such record:" + phone)
	}

	accounts := make([](*Account), len(rows))
	for idx := len(rows) - 1; idx > -1; idx-- {
		accounts[idx] = convRowMap2Account(rows[idx])
	}
	return accounts, nil
}

func GetAccountListByPhoneOrUID(condition string) ([](*Account), error) {
	if len(condition) == 0 || !isNum(condition) {
		return nil, errors.New("condition" + condition + " is Wrongful ")
	}
	sql := "select * from account where uid = ? union all select * from account where phone = ?"
	uid, phone := utils.Str2Int64(condition), condition

	rows := gDbUser.Query(sql, uid, phone)
	// logger.Info("GetAccountListByPhoneOrUID:--------------------------", rows, len(rows))
	if (rows == nil) || (len(rows) < 1) {
		return nil, errors.New("no such record:" + condition)
	}

	accounts := make([](*Account), len(rows))
	//for idx := len(rows) - 1; idx > -1; idx-- {
	//	accounts[idx] = convRowMap2Account(rows[idx])
	//}
	for i, v := range rows {
		accounts[i] = convRowMap2Account(v)
	}
	return accounts, nil
}

func SetEmail(uid int64, email string) error {
	_, err := gDbUser.Exec("update account set email = ? where uid = ?", email, uid)
	return err
}

func SetNickname(uid int64, nickname string) error {
	_, err := gDbUser.Exec("update account set nickname = ? where uid = ?", nickname, uid)
	return err
}

func SetPhone(uid int64, country int, phone string) error {
	_, err := gDbUser.Exec("update account set country = ?,phone = ? where uid = ?", country, phone, uid)
	return err
}
func SetLoginPassword(uid int64, password string) error {
	_, err := gDbUser.Exec("update account set login_password = ? where uid = ?", password, uid)
	return err
}

func SetPaymentPassword(uid int64, password string) error {
	_, err := gDbUser.Exec("update account set payment_password = ? where uid = ?", password, uid)
	return err
}

func SetAssetStatus(uid int64, status int) error {
	_, err := gDbUser.Exec("update account set status = ? where uid = ?", status, uid)
	return err
}

func SetUserLevel(uid int64, level int) error {
	_, err := gDbUser.Exec("update account set level = ? where uid = ?", level, uid)
	return err
}

func CheckLoginPwd(uid int64, pwdInDB string) bool {
	row, err := gDbUser.QueryRow("select login_password from account where uid = ? ", uid)
	if err != nil {
		logger.Error("query err ", err.Error())
		return false
	}
	pwd := utils.Sha256(pwdInDB + utils.Int642Str(uid))
	return pwd == row["login_password"]
}

func CheckPaymentPwd(uid int64, pwdInDB string) bool {
	row, err := gDbUser.QueryRow("select payment_password from account where uid = ? ", uid)
	if err != nil {
		logger.Error("query err ", err.Error())
		return false
	}
	pwd := utils.Sha256(pwdInDB + utils.Int642Str(uid))
	return pwd == row["payment_password"]
}

func convRowMap2Account(row map[string]string) *Account {
	if len(row) > 0 {
		account := &Account{}
		account.ID = utils.Str2Int64(row["id"])
		account.UID = utils.Str2Int64(row["uid"])
		account.UIDString = row["uid"]
		account.Nickname = row["nickname"]
		account.Email = row["email"]
		account.Country = utils.Str2Int(row["country"])
		account.Phone = row["phone"]
		account.Language = row["language"]
		account.Region = row["region"]
		account.From = row["from"]
		account.RegisterTime = utils.Str2Int64(row["register_time"])
		account.UpdateTime = utils.Str2Int64(row["update_time"])
		account.RegisterType = utils.Str2Int(row["register_type"])
		account.LoginPassword = row["login_password"]
		account.PaymentPassword = row["payment_password"]
		account.Level = utils.Str2Int(row["level"])
		account.Status = utils.Str2Int(row["status"])
		return account
	}
	return nil
}
func isNum(s string) bool {
	r, _ := regexp.Compile("[0-9]*")
	return r.MatchString(s)
}


func CheckWXIdExists(wxid string)bool{
	row,err := gDbUser.QueryRow("select count(1) as c from account_extend where wxid = ?",wxid)
	if err != nil {
		logger.Error("query account_extend error",err.Error())
		return false
	}
	return utils.Str2Int(row["c"])>0
}

func InitAccountExtend(uid int64)error{
	_, err := gDbUser.Exec("insert ignore into account_extend (uid,wx_openid,wx_unionid,tg_id,update_time) values (?,null,null,null,?)", uid, utils.GetTimestamp13())
	return err
}

func SetWxId(uid int64,wxOpenid,wxUnionid string) (int64,error) {
	sql := "update account_extend set wx_openid = ?,wx_unionid = ?,update_time = ? where uid = ? and wx_openid is null and wx_unionid is null"
	res, err := gDbUser.Exec(sql, wxOpenid,wxUnionid,utils.GetTimestamp13(), uid)
	if err != nil {
		return 0,err
	}
	return res.RowsAffected()
}

func SetTGId(uid int64,tgId string) (int64,error) {
	sql := "update account_extend set tg_id = ?,update_time = ? where uid = ? and tg_id is null"
	res, err := gDbUser.Exec(sql, tgId,utils.GetTimestamp13(), uid)
	if err != nil {
		return 0,err
	}
	return res.RowsAffected()
}


func CheckBindWx(uid int64)bool{
	sql := `
		select
			a.country as c,
			ae.wx_openid as openid,
			ae.wx_unionid as unionid
		from
			account as a
		left join
			account_extend as ae
		on
			a.uid = ae.uid
		where
			a.uid = ?
	`
	row,err := gDbUser.QueryRow(sql,uid)
	if err != nil {
		logger.Error("query bind info error",err.Error())
		return false
	}
	ac := utils.Str2Int(row["c"])
	if ac != 86 && ac != 852 && ac != 853 && ac != 886  {
		return true
	}
	if id,ok := row["openid"];ok{
		return len(id)>0
	}
	if id,ok := row["unionid"];ok{
		return len(id)>0
	}
	return false
}

func GetUserLevel(uid int64)int{
	row,err := gDbUser.QueryRow("select level from account where uid = ?",uid)
	if err != nil || row == nil{
		return -1
	}
	return utils.Str2Int(row["level"])
}

func CheckBindWXByUidAndCreditScore(uid int64,country int)(bool,bool,int){
	row,err := gDbUser.QueryRow("select wx_openid,wx_unionid,tg_id,credit_score from account_extend where uid = ?",uid)
	if err != nil {
		return false,false,0
	}
	if row == nil {
		InitAccountExtend(uid)
		return false,false,70
	}
	return len(row["wx_openid"])>0&&len(row["wx_unionid"])>0,len(row["tg_id"])>0,utils.Str2Int(row["credit_score"])
}

func GetUserCreditScore(uid int64)int{
	row,err := gDbUser.QueryRow("select credit_score from account_extend where uid = ?",uid)
	if err != nil {
		return 0
	}
	if row == nil {
		InitAccountExtend(uid)
		return 70
	}
	return utils.Str2Int(row["credit_score"])
}

func GetUserRegisterTime(uid int64)int64{
	row,err := gDbUser.QueryRow("select register_time from account where uid = ?",uid)
	if err != nil || row == nil {
		return 0
	}
	return utils.Str2Int64(row["register_time"])
}

func GetUserExtendByUid(uid int64)(string,string,int){
	row,err := gDbUser.QueryRow("select wx_openid,wx_unionid,tg_id,credit_score from account_extend where uid = ?",uid)
	if err != nil {
		return "","",0
	}
	if row == nil {
		InitAccountExtend(uid)
		return "","",70
	}
	return row["wx_openid"],row["wx_unionid"],utils.Str2Int(row["credit_score"])
}


func DeductionCreditScore(uid int64,score int)bool{
	_,err := gDbUser.Exec("update account_extend set credit_score = credit_score - ? where uid = ?",score,uid)
	if err != nil {
		logger.Error("DeductionCreditScore failed",err.Error())
		return false
	}
	return true
}

func GetUserWalletAddressByUid(uid int64) string {
	res, err := gDBAsset.QueryRow("select wallet_address from user_restrict where uid = ?", uid)
	if err != nil {
		logger.Error("cannot get wallet address ", err.Error())
		return ""
	}
	if res == nil {
		logger.Info("can not find wallet address by uid ", uid)
		return ""
	}
	return res["wallet_address"]
}

func SetWalletAddress(uid int64, walletAddress string) (int64, error) {
	//result, err := gDbUser.Exec("update account_extend set wallet_address = ? where uid = ? and (select count(1) from account_extend where wallet_address = ?) = 0", walletAddress, uid, walletAddress)
	result, err := gDbUser.Exec("update account_extend set wallet_address = ? where uid = ?", walletAddress, uid, walletAddress)
	if err != nil {
		logger.Error("exec sql error",err.Error())
		return 0,err
	}
	rowsAffected, _ := result.RowsAffected()
	return rowsAffected,err

}