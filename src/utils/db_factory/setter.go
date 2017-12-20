package db_factory

import (
	"database/sql"
	log "utils/logger"
)

func (m DBPool) Query(query string, args ...interface{}) []map[string]string {
	log.Debug("Query sql :(%s)", query)
	rows, err := m.currDB.Query(query, args...)
	defer rows.Close()
	if err != nil {
		log.Error("query error %s", err.Error())
		return nil
	}
	res := make([]map[string]string, 0)
	for rows.Next() {
		res = append(res, parseRow(rows))
	}
	return res
}

func (m DBPool) QueryRow(query string, args ...interface{}) map[string]string {
	log.Debug("Query Row sql :(%s)", query)
	rows, err := m.currDB.Query(query, args...)
	defer rows.Close()
	if err != nil {
		log.Error("query error %s", err.Error())
		return nil
	}
	if rows.Next() {
		return parseRow(rows)
	}
	return nil
}

func (m DBPool) Exec(sql string, args ...interface{}) (sql.Result,error) {
	log.Debug("Exec sql :(%s)", sql)
	res, err := m.currDB.Exec(sql, args...)
	if err != nil {
		log.Error("exec error %s", err.Error())
	}
	return res,err
}

func (m DBPool) Prepare(query string) (*sql.Stmt, error) {
	return m.currDB.Prepare(query)
}

func (m DBPool) GetDb() *sql.DB {
	return m.currDB
}

func (m DBPool) Close() {
	m.currDB.Close()
}

func (m DBPool) Ping() error {
	return m.currDB.Ping()
}

func (m DBPool) IsConn() bool {
	return m.isConn
}

func (m DBPool) Err() error {
	return m.err
}