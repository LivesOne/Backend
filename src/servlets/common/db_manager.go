package common

import (
	"database/sql"
	_ "fmt"
	_ "github.com/go-sql-driver/mysql"
	_ "utils/config"
	"utils/logger"
)

var gDbUser *sql.DB

func DbInit() error {
	/*
		dsn_ := fmt.Sprint("%s:%s@tcp(%s)/%s?charset=utf8",
			config.GetConfig().DBUser,
			config.GetConfig().DBUserPwd,
			config.GetConfig().DBHost,
			config.GetConfig().DBDatabase)
	*/

	dsn := "lvt_serv:U9D$WHTo@(10.100.15.96:3306)/livesone_user?charset=utf8"

	var err error
	gDbUser, err = sql.Open("mysql", dsn)
	if err != nil {
		logger.Fatal(err)
		return err
	}

	logger.Debug("connection database successful")
	return nil
}

func InsertAccount(account Account) (int64, error) {
	if gDbUser == nil {
		logger.Error("database not ready")
		return 0, nil
	}

	stmt, err := gDbUser.Prepare("INSERT account SET uid=?, email=?, country=?, phone=?, login_password=?, " +
		"`language`=?, region=?, `from`=?, register_time=?, update_time=?, register_type=?")
	if err != nil {
		logger.Fatal(err)
		return 0, err
	}

	res, err := stmt.Exec(account.UID, account.Email, account.Country, account.Phone, account.LoginPassword,
		account.Language, account.Region, account.From, account.RegisterTime, account.UpdateTime, account.RegisterType)
	if err != nil {
		logger.Fatal(err)
		return 0, err
	}

	id, _ := res.LastInsertId()

	return id, nil
}
