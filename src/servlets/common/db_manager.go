package common

import (
	"database/sql"
	"fmt"
	"utils/config"
	"utils/logger"
)

var gDbUser *sql.DB

func DbInit() error {
	dsn := fmt.Sprint("%s:%s@tcp(%s)/%s?charset=utf8",
		config.GetConfig().DBUser,
		config.GetConfig().DBUserPwd,
		config.GetConfig().DBHost,
		config.GetConfig().DBDatabase)

	var err error
	gDbUser, err = sql.Open("mysql", dsn)
	if err != nil {
		logger.Fatal(err)
		return err
	}
}

func InsertAccount() (int64, error) {

}
