package accounts

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"utils"
)

const (
	CHECK_TYPE_UID   = 1
	CHECK_TYPE_EMAIL = 2
	CHECK_TYPE_PHONE = 3

	CHECK_ACCOUNT_EXISTS     = 1
	CHECK_ACCOUNT_NOT_EXISTS = 2
)

type checkAccountRequest struct {
	Base  *common.BaseInfo   `json:"base"`
	Param *checkAccountParam `json:"param"`
}

type checkAccountParam struct {
	Type    int    `json:"type"`
	Country int    `json:"country"`
	Phone   string `json:"phone"`
	EMail   string `json:"email"`
	Uid     string `json:"uid"`
}

type checkAccountResponse struct {
	Exists int    `json:"exists"`
	Uid    string `json:"uid"`
	Status int    `json:"status"`
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
	resData := checkAccountResponse{Exists: CHECK_ACCOUNT_NOT_EXISTS}

	switch data.Param.Type {
	case CHECK_TYPE_UID:
		uid, status := common.GetAssetByUid((utils.Str2Int64(data.Param.Uid)))
		if uid != 0 {
			resData.Exists = CHECK_ACCOUNT_EXISTS
			resData.Uid = utils.Int642Str(uid)
			resData.Status = status
		}
	case CHECK_TYPE_EMAIL:
		uid, status := common.GetAssetByEmail(data.Param.EMail)
		if uid != 0 {
			resData.Exists = CHECK_ACCOUNT_EXISTS
			resData.Uid = utils.Int642Str(uid)
			resData.Status = status
		}
	case CHECK_TYPE_PHONE:
		uid, status := common.GetAssetByPhone(data.Param.Country, data.Param.Phone)
		if uid != 0 {
			resData.Exists = CHECK_ACCOUNT_EXISTS
			resData.Uid = utils.Int642Str(uid)
			resData.Status = status
		}
	}
	response.Data = resData
}
