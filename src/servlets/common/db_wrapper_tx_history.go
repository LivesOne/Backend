package common

type (
	DTTXHistory struct {
		Id      int64       `bson:"_id"`
		Status  int         `bson:"status"`
		Type    int         `bson:"type"`
		TradeNo string      `bson:"trade_no"`
		From    int64       `bson:"from,omitempty"`
		To      int64       `bson:"to,omitempty"`
		Value   int64       `bson:"value"`
		Ts      int64       `bson:"ts"`
		Code    int         `bson:"code"`
		Remark  interface{} `bson:"remark"`
		Miner   []Miner     `bson:"miner,omitempty"`
	}

	Miner struct {
		Sid   int   `bson:"sid" json:"sid"`
		Value int64 `bson:"value" json:"value"`
	}
)
