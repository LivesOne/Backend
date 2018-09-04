package common

import (
	"utils"
	"utils/logger"
)

const (
	REDIS_TRADE_KEY_PROXY = "pay:order:"
)

type (
	TradeMiner struct {
		Sid   int   `json:"sid,omitempty" bson:"sid,omitempty"`
		Value int64 `json:"value,omitempty" bson:"value,omitempty"`
	}

	TradePay struct {
		AppId          string `json:"app_id,omitempty" bson:"app_id,omitempty"`
		ReturnUrl      string `json:"return_url,omitempty" bson:"return_url,omitempty"`
		NotifyUrl      string `json:"notify_url,omitempty" bson:"notify_url,omitempty"`
		Body           string `json:"body,omitempty" bson:"body,omitempty"`
		TimeoutExpired string `json:"timeout_expired,omitempty" bson:"timeout_expired,omitempty"`
	}

	TradeConversion struct {
		OriginalCurrency string `json:"original_currency" bson:"original_currency,omitempty"`
		TargetCurrency   string `json:"target_currency" bson:"target_currency,omitempty"`
	}

	TradeRecharge struct {
		Hash    string `json:"hash,omitempty" bson:"hash,omitempty"`
		Address string `json:"address,omitempty" bson:"address,omitempty"`
	}

	TradeWithdrawal struct {
		Hash    string `json:"hash,omitempty" bson:"hash,omitempty"`
		Address string `json:"address,omitempty" bson:"address,omitempty"`
	}

	TradeInfo struct {
		TradeNo         string           `json:"trade_no,omitempty" bson:"trade_no,omitempty"`
		Type            int              `json:"type,omitempty" bson:"type,omitempty"`
		SubType         int              `json:"sub_type,omitempty" bson:"sub_type,omitempty"`
		From            int64            `json:"from,omitempty" bson:"from,omitempty"`
		FromName        int64            `json:"from_name,omitempty" bson:"from_name,omitempty"`
		To              int64            `json:"to,omitempty" bson:"to,omitempty"`
		ToName          int64            `json:"to_name,omitempty" bson:"to_name,omitempty"`
		Amount          int64            `json:"amount,omitempty" bson:"amount,omitempty"`
		Decimal         int              `json:"decimal,omitempty" bson:"decimal,omitempty"`
		Currency        string           `json:"currency,omitempty" bson:"currency,omitempty"`
		CreateTime      int64            `json:"create_time,omitempty" bson:"create_time,omitempty"`
		Status          int              `json:"status" bson:"status"`
		Subject         string           `json:"subject,omitempty" bson:"subject,omitempty"`
		Txid            int64            `json:"txid,omitempty" bson:"txid,omitempty"`
		OutTradeNo      string           `json:"out_trade_no,omitempty" bson:"out_trade_no,omitempty"`
		OutRequestNo    string           `json:"out_request_no,omitempty" bson:"out_request_no,omitempty"`
		RefundTradeNo   string           `json:"refund_trade_no,omitempty" bson:"refund_trade_no,omitempty"`
		OriginalTradeNo string           `json:"original_trade_no,omitempty" bson:"original_trade_no,omitempty"`
		FeeTradeNo      string           `json:"fee_trade_no,omitempty" bson:"fee_trade_no,omitempty"`
		FinishTime      int64            `json:"finish_time,omitempty" bson:"finish_time,omitempty"`
		Miner           []TradeMiner     `json:"miner,omitempty" bson:"miner,omitempty"`
		Pay             *TradePay        `json:"pay,omitempty" bson:"pay,omitempty"`
		Conversion      *TradeConversion `json:"conversion,omitempty" bson:"conversion,omitempty"`
		Recharge        *TradeRecharge   `json:"recharge,omitempty" bson:"recharge,omitempty"`
		Withdrawal      *TradeWithdrawal `json:"withdrawal,omitempty" bson:"withdrawal,omitempty"`
	}
)

func (tradeInfo *TradeInfo) TryLock(value int64) bool {
	nowTimestamp := utils.GetTimestamp13()
	key := REDIS_TRADE_KEY_PROXY + tradeInfo.TradeNo
	nx, err := setnx(key, nowTimestamp)
	if err != nil {
		logger.Error("lock the order error for pay, key:", key)
		return false
	}
	if nx == 1 {
		rdsExpire(key, 300)
	}
	return nx == 1
}

func (tradeInfo *TradeInfo) ReleaseLock(value int64) bool {
	key := REDIS_TRADE_KEY_PROXY + tradeInfo.TradeNo
	rdsValue, _ := rdsGet64(key)
	if rdsValue == value {
		err := rdsDel(key)
		if err != nil {
			logger.Error("release order share lock error, error:", err.Error())
			return false
		} else {
			return true
		}
	} else {
		return true
	}
}
