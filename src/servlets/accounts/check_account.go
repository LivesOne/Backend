package accounts

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
)

const (
	CHECK_TYPE_UID = 1
	CHECK_TYPE_EMAIL = 2
	CHECK_TYPE_PHONE = 3

)
type checkAccountRequest struct {
	Base  common.BaseInfo   `json:"base"`
	Param checkAccountParam `json:"param"`
}

type checkAccountParam struct {
	Type    int    `json:"type"`
	Country int    `json:"country"`
	Phone   string `json:"phone"`
	EMail   string `json:"email"`
	Uid     int64    `json:"uid"`
}

type checkAccountResponse struct {
	Exists int `json:"exists"`
	Uid int64 `json:"uid"`
}
// checkVCodeHandler
type checkAccountHandler struct {
	//header      *common.HeaderParams // request header param
	//requestData *checkVCodeRequest   // request body
}

func (handler *checkAccountHandler) Method() string {
	return http.MethodPost
}

func (handler *checkAccountHandler) Handle(request *http.Request, writer http.ResponseWriter) {

	response := &common.ResponseData{
		Base: &common.BaseResp{
			RC:  constants.RC_OK.Rc,
			Msg: constants.RC_OK.Msg,
		},
	}
	defer common.FlushJSONData2Client(response, writer)
	data := checkAccountRequest{}
	//header := common.ParseHttpHeaderParams(request)
	common.ParseHttpBodyParams(request, &data)
	resData := checkAccountResponse{Exists:2}

	switch data.Param.Type {
	case CHECK_TYPE_UID:
		if common.ExistsUID(data.Param.Uid) {
			resData.Exists = 1
			resData.Uid = data.Param.Uid
		}
	case CHECK_TYPE_EMAIL:
		uid := common.GetUidByEmail(data.Param.EMail)
		if uid != 0 {
			resData.Exists = 1
			resData.Uid = uid
		}
	case CHECK_TYPE_PHONE:
		uid := common.GetUidByPhone(data.Param.Country,data.Param.Phone)
		if uid != 0 {
			resData.Exists = 1
			resData.Uid = uid
		}
	}
	response.Data = resData
}
