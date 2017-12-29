package common

type (
	Reward struct {
		Total      int64 `json:"total"`
		Yesterday  int64 `json:"yesterday"`
		Lastmodify int64 `json:"lastmodify"`
		Uid        int64 `json:"uid"`
	}
)
