package common

import (
	"strings"
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


func GetWithdrawQuotaByCurrency(currency string) *WithdrawQuota {
	sql := "select * from dt_withdrawal_amount where currency = ?"
	row, err := gDBConfig.QueryRow(sql, currency)
	if err != nil {
		logger.Error("query quota from dt_withdrawal_amount by currency error, currency:", currency, ", error:", err.Error())
	}
	withdrawQuota := WithdrawQuota{
		Currency:        currency,
		SingleAmountMin: utils.Str2Float64(row["single_amount_min"]),
		DailyAmountMax:  utils.Str2Float64(row["daily_amount_max"]),
		UpdateTime:      utils.Str2Int64(row["update_time"]),
	}

	withdrawFeeArray := make([]WithdrawFee, 0)
	sql = "select * from dt_withdrawal_fee where currency = ?"
	rows := gDBConfig.Query(sql, currency)
	for _, row = range rows {
		withdrawFee := WithdrawFee{
			Id:          utils.Str2Int64(row["id"]),
			Currency:    row["currency"],
			FeeCurrency: row["fee_currency"],
			FeeType:     utils.Str2Int(row["fee_type"]),
			FeeFixed:    utils.Str2Float64(row["fee_fixed"]),
			FeeRate:     utils.Str2Float64(row["fee_rate"]),
			FeeMin:      utils.Str2Float64(row["fee_min"]),
			FeeMax:      utils.Str2Float64(row["fee_max"]),
			Discount:    utils.Str2Float64(row["discount"]),
			UpdateTime:  utils.Str2Int64(row["update_time"]),
		}
		withdrawFeeArray = append(withdrawFeeArray, withdrawFee)
	}
	withdrawQuota.Fee = withdrawFeeArray
	return &withdrawQuota
}

func QueryCurrencyPrice(currency, currency2 string) (string, string, error) {
	row, err := gDBConfig.QueryRow(
		"select format(current,8) as cur,format(average,8) as avg from dt_currency_price where currency=? and currency2 = ?",
		currency, currency2)
	if err != nil {
		logger.Warn("query currency price error:", err)
		return "", "", err
	}
	if row != nil {
		return row["cur"], row["avg"], nil
	}
	logger.Info("currency price not found:", currency, ",", currency2)
	return "", "", nil
}

func GeTransferQuotaByCurrency(currency, feeCurrency string) *TransferQuota {
	sql := "select * from dt_transfer_amount where currency = ?"
	row, err := gDBConfig.QueryRow(sql, currency)
	if err != nil {
		logger.Error("query quota from dt_transfer_amount by currency error, currency:", currency, ", error:", err.Error())
	}
	transferQuota := TransferQuota{
		Currency:        currency,
		SingleAmountMin: utils.Str2Float64(row["single_amount_min"]),
		DailyAmountMax:  utils.Str2Float64(row["daily_amount_max"]),
		UpdateTime:      utils.Str2Int64(row["update_time"]),
	}

	sql = "select * from dt_transfer_fee where currency = ? and fee_currency = ?"
	row, err = gDBConfig.QueryRow(sql, currency, feeCurrency)
	transferFee := TransferFee{
		Id:          utils.Str2Int64(row["id"]),
		Currency:    row["currency"],
		FeeCurrency: row["fee_currency"],
		FeeType:     utils.Str2Int(row["fee_type"]),
		FeeFixed:    utils.Str2Float64(row["fee_fixed"]),
		FeeRate:     utils.Str2Float64(row["fee_rate"]),
		FeeMin:      utils.Str2Float64(row["fee_min"]),
		FeeMax:      utils.Str2Float64(row["fee_max"]),
		Discount:    utils.Str2Float64(row["discount"]),
		UpdateTime:  utils.Str2Int64(row["update_time"]),
	}
	transferQuota.Fee = transferFee
	return &transferQuota
}

func GetFeeCurrencyByCurrency(currency string) (string, error) {
	sql := "select fee_currency from dt_withdrawal_fee where currency = ? limit 1"
	row, err := gDBConfig.QueryRow(sql, strings.ToUpper(currency))
	if err != nil {
		logger.Error("get withdraw fee currency by currency err, currency:", strings.ToUpper(currency))
		return "", err
	}
	return row["fee_currency"], nil
}

func ConversionCoinPrice(amount float64, source, target string) float64 {
	sql := "select average from dt_currency_price where currency = ? and currency2 = ?"
	row, err := gDBConfig.QueryRow(sql, strings.ToUpper(source), strings.ToUpper(target))
	if err != nil {
		logger.Error("get price by currency err, currency:", strings.ToUpper(source), "currency2:", strings.ToUpper(target), "err:", err)
	}
	if row != nil && len(row) > 0 {
		average := utils.Str2Float64(row["average"])
		return amount * average
	}

	sql = "select average from dt_currency_price where currency = ? and currency2 = ?"
	row, err = gDBConfig.QueryRow(sql, strings.ToUpper(target), strings.ToUpper(source))
	if err != nil {
		logger.Error("get price by currency err, currency:", strings.ToUpper(target), "currency2:", strings.ToUpper(source), "err:", err)
	}
	if row != nil && len(row) > 0 {
		average := utils.Str2Float64(row["average"])
		return amount / average
	}

	average1, average2 := float64(0), float64(0)
	sql = "select average from dt_currency_price where currency = ? and currency2 = ?"
	row, err = gDBConfig.QueryRow(sql, strings.ToUpper(source), "USDT")
	if err != nil {
		logger.Error("get price by currency err, currency:", strings.ToUpper(source), "currency2: USDT", "err:", err)
		return float64(-1)
	}
	if row != nil && len(row) > 0 {
		average1 = utils.Str2Float64(row["average"])
	} else {
		return float64(-1)
	}
	sql = "select average from dt_currency_price where currency = ? and currency2 = ?"
	row, err = gDBConfig.QueryRow(sql, strings.ToUpper(target), "USDT")
	if err != nil {
		logger.Error("get price by currency err, currency:", strings.ToUpper(target), "currency2: USDT", "err:", err)
		return float64(-1)
	}
	if row != nil && len(row) > 0 {
		average2 = utils.Str2Float64(row["average"])
	} else {
		return float64(-1)
	}
	return amount * average1 / average2
}