package db_factory

import (
	"database/sql"
	log "utils/logger"
	"time"
	"strconv"
)

func (m *DBPool) Query(query string, args ...interface{}) []map[string]string {
	start := time.Now()
	defer logTime(start,query)
	log.Debug("Query sql :(", query,")")
	rows, err := m.currDB.Query(query, args...)
	defer rows.Close()
	if err != nil {
		log.Error("query error ", err.Error())
		return nil
	}
	res := make([]map[string]string, 0)
	for rows.Next() {
		res = append(res, parseRow(rows))
	}
	return res
}

func (m *DBPool) QueryRow(query string, args ...interface{}) (map[string]string,error) {
	start := time.Now()
	defer logTime(start,query)
	log.Debug("Query Row sql :(", query,")")
	rows, err := m.currDB.Query(query, args...)
	defer rows.Close()
	if err != nil {
		log.Error("query error ", err.Error())
	}
	if rows.Next() {
		return parseRow(rows),nil
	}
	return nil,err
}

func (m *DBPool) Exec(sql string, args ...interface{}) (sql.Result,error) {
	start := time.Now()
	defer logTime(start,sql)
	log.Debug("Exec sql :(", sql,")")
	res, err := m.currDB.Exec(sql, args...)
	if err != nil {
		log.Error("exec error ", err.Error())
	}
	return res,err
}

func (m *DBPool) Prepare(query string) (*sql.Stmt, error) {
	return m.currDB.Prepare(query)
}

func (m *DBPool) GetDb() *sql.DB {
	return m.currDB
}

func (m *DBPool) Close() {
	m.currDB.Close()
}

func (m *DBPool) Ping() error {
	return m.currDB.Ping()
}

func (m *DBPool) IsConn() bool {
	return m.isConn
}

func (m *DBPool) Err() error {
	return m.err
}

func (m *DBPool) Begin()(*sql.Tx, error){
	return m.currDB.Begin()
}

func logTime(start time.Time, msg ...interface{}) {
	subSec := time.Now().UTC().Sub(start).Seconds()
	ms := subSec * 1000
	dis := strconv.FormatFloat(ms, 'f', 2, 64)
	log.Debug(msg, " ", dis, " ms")
}