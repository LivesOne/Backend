package common

import (
	"database/sql"
	"fmt"
	_ "fmt"
	_ "github.com/go-sql-driver/mysql"
	"strconv"
	"utils/config"
	_ "utils/config"
	"utils/db_factory"
	"utils/logger"
)

//var gDbUser *sql.DB
var gDbUser db_factory.DBPool

func DbInit() error {

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
	//	logger.Fatal(err)
	//	return err
	//}
	db_config := config.GetConfig()
	facConfig := db_factory.Config{
		Host:        db_config.DBHost,
		UserName:    db_config.DBUser,
		Password:    db_config.DBUserPwd,
		Database:    db_config.DBDatabase,
		MaxConn:     10,
		MaxIdleConn: 1,
	}
	gDbUser = db_factory.NewDataSource(facConfig)
	if gDbUser.IsConn() {
		logger.Debug("connection database successful")
	} else {
		logger.Fatal(gDbUser.Err())
	}

	return nil
}

func ExistsUID(uid int64) bool {
	row := gDbUser.QueryRow("select count(1) as c from account where uid = ? limit 1", uid)
	i, _ := strconv.Atoi(row["c"])
	return i > 0
}

func ExistsEmail(email string) bool {
	row := gDbUser.QueryRow("select count(1) as c from account where email = ? limit 1", email)
	i, _ := strconv.Atoi(row["c"])
	return i > 0
}

func ExistsPhone(country int, phone string) bool {
	row := gDbUser.QueryRow("select count(1) as c from account where country = ? and phone = ? limit 1", country, phone)
	i, _ := strconv.Atoi(row["c"])
	return i > 0
}

func InsertAccount(account Account) (int64, error) {
	if !gDbUser.IsConn() {
		logger.Error("database not ready")
		return 0, nil
	}
	stmt, err := gDbUser.Prepare("INSERT account SET uid=?, login_password=?, " +
		"`language`=?, region=?, `from`=?, register_time=?, update_time=?, register_type=?")
	if err != nil {
		logger.Fatal(err)
		return 0, err
	}

	ret, err := stmt.Exec(account.UID, account.LoginPassword,
		account.Language, account.Region, account.From, account.RegisterTime, account.UpdateTime, account.RegisterType)
	if err != nil {
		logger.Fatal(err)
		return 0, err
	}

	id, _ := ret.LastInsertId()

	return id, nil
}

func InsertAccountWithEmail(account Account) (int64, error) {
	if !gDbUser.IsConn() {
		logger.Error("database not ready")
		return 0, nil
	}

	stmt, err := gDbUser.Prepare("INSERT account SET uid=?, email=?, login_password=?, " +
		"`language`=?, region=?, `from`=?, register_time=?, update_time=?, register_type=?")
	if err != nil {
		logger.Fatal(err)
		return 0, err
	}

	ret, err := stmt.Exec(account.UID, account.Email, account.LoginPassword,
		account.Language, account.Region, account.From, account.RegisterTime, account.UpdateTime, account.RegisterType)
	if err != nil {
		logger.Fatal(err)
		return 0, err
	}

	id, _ := ret.LastInsertId()

	return id, nil
}

func InsertAccountWithPhone(account Account) (int64, error) {
	if !gDbUser.IsConn() {
		logger.Error("database not ready")
		return 0, nil
	}

	stmt, err := gDbUser.Prepare("INSERT account SET uid=?, country=?, phone=?, login_password=?, " +
		"`language`=?, region=?, `from`=?, register_time=?, update_time=?, register_type=?")
	if err != nil {
		logger.Fatal(err)
		return 0, err
	}

	ret, err := stmt.Exec(account.UID, account.Country, account.Phone, account.LoginPassword,
		account.Language, account.Region, account.From, account.RegisterTime, account.UpdateTime, account.RegisterType)
	if err != nil {
		logger.Fatal(err)
		return 0, err
	}

	id, _ := ret.LastInsertId()

	return id, nil
}

func GetAccountByUID(uid string) (Account, error) {
	var account Account

	return account, nil
}

func GetAccountByEmail(email string) (Account, error) {
	var account Account

	return account, nil
}

func GetAccountByPhone(country int, phone string) (Account, error) {
	var account Account

	return account, nil
}

func SetEmail(uid int64, email string) error {
	_,err := gDbUser.Exec("update account set email = ? where uid = ?",email,uid)
	return err
}

func SetPhone(uid int64, country int, phone string) error {
	_,err := gDbUser.Exec("update account set country = ?,phone = ? where uid = ?",country,phone,uid)
	return err
}
func SetLoginPassword(uid int64, password string) error {
	_,err := gDbUser.Exec("update account set login_password = ? where uid = ?",password,uid)
	return err
}

func SetPaymentPassword(uid int64, password string) error {
	_,err := gDbUser.Exec("update account set payment_password = ? where uid = ?",password,uid)
	return err
}
