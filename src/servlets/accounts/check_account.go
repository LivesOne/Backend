package accounts

import (
	"gitlab.maxthon.net/cloud/livesone-user-micro/src/proto"
	"golang.org/x/net/context"
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/rpc"
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
	Status int64  `json:"status"`
}

// checkVCodeHandler
type checkAccountHandler struct {
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

	if data.Param == nil {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	resData := checkAccountResponse{Exists: CHECK_ACCOUNT_NOT_EXISTS}

	cli := rpc.GetUserCacheClient()
	if cli == nil {
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}

	switch data.Param.Type {
	case CHECK_TYPE_UID:
		req := &microuser.UserIdReq{
			Uid: utils.Str2Int64(data.Param.Uid),
		}
		resp, err := cli.CheckAccountByUid(context.Background(), req)
		if err != nil {
			response.SetResponseBase(constants.RC_SYSTEM_ERR)
			return
		}
		if resp.Result == microuser.ResCode_OK {
			resData.Exists = CHECK_ACCOUNT_EXISTS
			resData.Uid = utils.Int642Str(resp.Uid)
			resData.Status = resp.Status
		}
	case CHECK_TYPE_EMAIL:
		req := &microuser.CheckAccountByEmailReq{
			Email: data.Param.EMail,
		}
		resp, err := cli.CheckAccountByEmail(context.Background(), req)
		if err != nil {
			response.SetResponseBase(constants.RC_SYSTEM_ERR)
			return
		}
		if resp.Result == microuser.ResCode_OK {
			resData.Exists = CHECK_ACCOUNT_EXISTS
			resData.Uid = utils.Int642Str(resp.Uid)
			resData.Status = resp.Status
		}
	case CHECK_TYPE_PHONE:
		req := &microuser.CheckAccountByPhoneReq{
			Country: int64(data.Param.Country),
			Phone:   data.Param.Phone,
		}
		resp, err := cli.CheckAccountByPhone(context.Background(), req)
		if err != nil {
			response.SetResponseBase(constants.RC_SYSTEM_ERR)
			return
		}
		if resp.Result == microuser.ResCode_OK {
			resData.Exists = CHECK_ACCOUNT_EXISTS
			resData.Uid = utils.Int642Str(resp.Uid)
			resData.Status = resp.Status
		}
	}
	response.Data = resData
}
