package common

import (
	"utils"
	"utils/config"
	"utils/db_factory"
	"utils/logger"
)

var gDBConfig *db_factory.DBPool

func ConfigDbInit() error {
	db_config := config.GetConfig().Config
	facConfig := db_factory.Config{
		Host:        db_config.DBHost,
		UserName:    db_config.DBUser,
		Password:    db_config.DBUserPwd,
		Database:    db_config.DBDatabase,
		MaxConn:     db_config.MaxConn,
		MaxIdleConn: db_config.MaxConn,
	}
	gDBConfig = db_factory.NewDataSource(facConfig)
	if gDBConfig.IsConn() {
		logger.Debug("connection ", db_config.DBHost, db_config.DBDatabase, "database successful")
	} else {
		logger.Error(gDBConfig.Err())
		return gDBConfig.Err()
	}
	return nil
}

func QueryTransFee(currency, feeCurrency string) (*DtTransferFee, error) {
	row, err := gDbUser.QueryRow(`select * 
		from account 
		where currency = ? and fee_currency = ?`,
		currency, feeCurrency)
	if err != nil {
		logger.Error("query dt_transfer_fee err,", err.Error())
		return nil, err
	}
	return convRowMap2DtTransFee(row), err
}

func convRowMap2DtTransFee(row map[string]string) *DtTransferFee {
	if row != nil && len(row) > 0 {
		transfee := &DtTransferFee{
			Currency: row["currency"],
			FeeCurrency: row["fee_currency"],
			FeeRate: utils.Str2Float64(row["fee_rate"]),
			Discount: utils.Str2Float64(row["discount"]),
			FeeMax: utils.Str2Int(row["fee_max"]),
			UpdateTime: utils.Str2Int64(row["update_time"]),
		}
		return transfee
	}
	return nil
}
