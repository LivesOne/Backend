package DBFactory

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	log "utils/logger"
	"fmt"
)




type DBConfig struct{
	User_name string
	Password string
	Host string
	Port string
	Db_name string
	MaxConn int
	MaxIdleConn int
}

type SingleDB struct{
	currDB *sql.DB
	isConn bool
}

func InitDbConfig(dbConfig DBConfig)(SingleDB){
	singleDB := initDb(dbConfig)
	err := singleDB.Ping()
	if err != nil {
		log.Error("cannot conn db %s",err.Error())
	}
	return singleDB
}

func convConfig2Str(dbConfig DBConfig)(string){
	connstr := "%s:%s@tcp(%s:%s)/%s?charset=utf8"
	return fmt.Sprintf(connstr,dbConfig.User_name,dbConfig.Password,dbConfig.Host,dbConfig.Port,dbConfig.Db_name)
}

func initDb(dbConfig DBConfig)(SingleDB){
	conn_str := convConfig2Str(dbConfig)
	db, err := sql.Open("mysql", conn_str)
	if err != nil {
		log.Error("openstr %s",conn_str)
		log.Error("cannot conn db %s",err.Error())
		return SingleDB{
			isConn:false,
		}
	}else{
		db.SetMaxOpenConns(dbConfig.MaxConn)
		db.SetMaxIdleConns(dbConfig.MaxIdleConn)
		return SingleDB{
			currDB:db,
			isConn:true,
		}
	}
}



func parseRow(r *sql.Rows)map[string]string{
	cols,_ := r.ColumnTypes() // Remember to check err afterwards
	values := make([]sql.RawBytes, len(cols))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	err := r.Scan(scanArgs...)
	if err != nil {
		log.Error("query error %s",err.Error())
	}
	var item map[string]string = make(map[string]string ,0)
	for i, col := range values {
		if col != nil {
			k := cols[i].Name()
			item[k] = string(col)
		}
	}
	return item
}
