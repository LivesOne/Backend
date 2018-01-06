package common

type (
	DTTXHistory struct {
		Id     int64   `bson:"_id"`
		Status int     `bson:"status"`
		Type   int     `bson:"type"`
		From   int64   `bson:"from"`
		To     int64   `bson:"to"`
		Value  int64   `bson:"value"`
		Ts     int64   `bson:"ts"`
		Code   int     `bson:"code"`
	}

)
