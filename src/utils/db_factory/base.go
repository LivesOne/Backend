package db_factory

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	log "utils/logger"
)

type Config struct {
	Host        string
	UserName    string
	Password    string
	Database    string
	MaxConn     int
	MaxIdleConn int
}

type DBPool struct {
	currDB *sql.DB
	isConn bool
	err error
}

func NewDataSource(config Config) DBPool {
	dbPool := initDb(config)
	err := dbPool.Ping()
	if err != nil {
		log.Error("cannot conn db %s", err.Error())
		dbPool.err = err
	}
	return dbPool
}

func convConfig2Str(config Config) string {
	dsn := "%s:%s@tcp(%s)/%s?charset=utf8"
	return fmt.Sprintf(dsn, config.UserName, config.Password, config.Host, config.Database)
}

func initDb(config Config) DBPool {
	connStr := convConfig2Str(config)
	db, err := sql.Open("mysql", connStr)
	if err != nil {
		log.Error("openstr %s", connStr)
		log.Error("cannot conn db %s", err.Error())
		return DBPool{
			isConn: false,
		}
	} else {
		db.SetMaxOpenConns(config.MaxConn)
		db.SetMaxIdleConns(config.MaxIdleConn)
		return DBPool{
			currDB: db,
			isConn: true,
		}
	}
}

func parseRow(r *sql.Rows) map[string]string {
	cols, _ := r.ColumnTypes() // Remember to check err afterwards
	values := make([]sql.RawBytes, len(cols))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	err := r.Scan(scanArgs...)
	if err != nil {
		log.Error("query error %s", err.Error())
	}
	var item map[string]string = make(map[string]string, 0)
	for i, col := range values {
		if col != nil {
			k := cols[i].Name()
			item[k] = string(col)
		}
	}
	return item
}
