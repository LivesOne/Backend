package common

import (
	"database/sql"
	_ "fmt"
	_ "github.com/go-sql-driver/mysql"
	_ "utils/config"
	"utils/logger"
	"fmt"
	"utils/config"
)

var gDbUser *sql.DB

func DbInit() error {

		dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8",
			config.GetConfig().DBUser,
			config.GetConfig().DBUserPwd,
			config.GetConfig().DBHost,
			config.GetConfig().DBDatabase)

	//dsn := "lvt_serv:U9D$WHTo@(10.100.15.96:3306)/livesone_user?charset=utf8"

	var err error
	gDbUser, err = sql.Open("mysql", dsn)
	if err != nil {
		logger.Fatal(err)
		return err
	}

	logger.Debug("connection database successful")
	return nil
}

func ExistsUID(uid int64) bool {

	return false
}

func ExistsEmail(email string) bool {

	return false
}

func ExistsPhone(country int, phone string) bool {

	return false
}

func InsertAccount(account Account) (int64, error) {
	if gDbUser == nil {
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
	if gDbUser == nil {
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
	if gDbUser == nil {
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

func GetAccountByEmail(email string) (Account, error)  {
	var account Account

	return account, nil
}

func GetAccountByPhone(country int, phone string) (Account, error)  {
	var account Account


	return account, nil
}

func SetEmail(uid int64, email string) error {

	return nil
}

func SetPhone(uid int64, country int, phone string) error {

	return nil
}
func SetLoginPassword(uid int64, password string)  error {

	return nil
}

func SetPaymentPassword(uid int64, password string) error {

	return nil
}
