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

func NewDataSource(config Config) *DBPool {
	return initDb(config)
}

func convConfig2Str(config Config) string {
	dsn := "%s:%s@tcp(%s)/%s?charset=utf8"
	return fmt.Sprintf(dsn, config.UserName, config.Password, config.Host, config.Database)
}

func initDb(config Config) *DBPool {
	connStr := convConfig2Str(config)
	db, err := sql.Open("mysql", connStr)
	if err != nil {
		log.Error("openstr", connStr)
		log.Error("cannot conn db", err.Error())
		return &DBPool{
			isConn: false,
			err:err,
		}
	} else {
		log.Info("init db conn pool --->",connStr)
		db.SetMaxOpenConns(config.MaxConn)
		db.SetMaxIdleConns(config.MaxIdleConn)
		return &DBPool{
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

func parseErrCode(err error) int{
	//TODO parse err to err CODE
	return ER_ABORTING_CONNECTION
}