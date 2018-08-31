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
	row, err := gDBConfig.QueryRow(`select * 
		from dt_transfer_fee 
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
			FeeMin: utils.Str2Float64(row["fee_min"]),
			FeeMax: utils.Str2Float64(row["fee_max"]),
			UpdateTime: utils.Str2Int64(row["update_time"]),
		}
		return transfee
	}
	return nil
}

func QueryTransAmount(currency string) (*DtTransferAmount, error) {
	row, err := gDBConfig.QueryRow(`select * 
		from dt_transfer_amount 
		where currency = ?`,
		currency)
	if err != nil {
		logger.Error("query dt_transfer_amount err,", err.Error())
		return nil, err
	}
	return convRowMap2DtTransAmount(row), err
}

func convRowMap2DtTransAmount(row map[string]string) *DtTransferAmount {
	if row != nil && len(row) > 0 {
		transAmount := &DtTransferAmount{
			Currency: row["currency"],
			SingleAmountMin: utils.Str2Float64(row["single_amount_min"]),
			DailyAmountMax: utils.Str2Float64(row["daily_amount_max"]),
			UpdateTime: utils.Str2Int64(row["update_time"]),
		}
		return transAmount
	}
	return nil
}

func QueryTransSingleAmountMin(currency string) (float64, error) {
	dtAmount, err := QueryTransAmount(currency)
	if err != nil {
		return 0, err
	}
	return dtAmount.SingleAmountMin, nil
}

func QueryTransDailyAmountMax(currency string) (float64, error) {
	dtAmount, err := QueryTransAmount(currency)
	if err != nil {
		return 0, err
	}
	return dtAmount.DailyAmountMax, nil
}

