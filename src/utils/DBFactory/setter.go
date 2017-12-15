package DBFactory

import (
	"database/sql"
	log "utils/logger"
)


func (m SingleDB)Query(query string, args ...interface{})([]map[string]string){

	log.Debug("Query sql :(%s)",query);
	rows ,err := m.currDB.Query(query,args...)
	defer rows.Close()
	if err != nil {
		log.Error("query error %s",err.Error())
		return nil;
	}
	res := make([]map[string]string, 0);
	for rows.Next() {
		res = append(res,parseRow(rows));
	}
	return res
}

func (m SingleDB)QueryRow(query string,args ...interface{})map[string]string{
	log.Debug("Query Row sql :(%s)",query);
	rows ,err := m.currDB.Query(query,args...)
	defer rows.Close()
	if err != nil {
		log.Error("query error %s",err.Error())
		return nil;
	}
	if rows.Next(){
		return parseRow(rows)
	}
	return nil
}

func (m SingleDB)Exec(query string, args ...interface{})(sql.Result){
	log.Debug("Exec sql :(%s)",query);
	res,err := m.currDB.Exec(query, args...)
	if err!= nil{
		log.Error("exec error %s",err.Error())
		return nil
	}
	return res
}

func (m SingleDB)Prepare(query string)(*sql.Stmt, error){
	return m.currDB.Prepare(query)
}

func(m SingleDB) GetDb()(*sql.DB){
	return m.currDB
}

func (m SingleDB)Close(){
	m.currDB.Close();
}

func (m SingleDB) Ping()error{
	return m.currDB.Ping()
}

func (m SingleDB) IsConn()bool{
	return m.isConn
}