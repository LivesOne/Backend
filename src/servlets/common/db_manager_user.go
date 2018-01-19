package common

import (
	"errors"
	_ "fmt"
	"utils/config"
	_ "utils/config"
	"utils/db_factory"
	"utils/logger"

	"utils"

	_ "github.com/go-sql-driver/mysql"
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
		logger.Debug("connection database successful")
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

func ExistsPhone(country int, phone string) bool {
	row, _ := gDbUser.QueryRow("select count(1) as c from account where country = ? and phone = ? limit 1", country, phone)
	if row == nil {
		return false
	}
	return utils.Str2Int(row["c"]) > 0
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
	}
	if len(row) < 1 {
		return nil, errors.New("no record for:" + uid)
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
	rows := gDbUser.Query("select * from account where phone = ? or uid = ? ", condition, condition)
	// logger.Info("GetAccountListByPhoneOrUID:--------------------------", rows, len(rows))
	if (rows == nil) || (len(rows) < 1) {
		return nil, errors.New("no such record:" + condition)
	}

	accounts := make([](*Account), len(rows))
	for idx := len(rows) - 1; idx > -1; idx-- {
		accounts[idx] = convRowMap2Account(rows[idx])
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
		var account *Account = &Account{}
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
		return account
	}
	return nil
}
