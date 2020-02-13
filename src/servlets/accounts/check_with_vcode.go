package accounts

import (
	"gitlab.maxthon.net/cloud/livesone-user-micro/src/proto"
	"golang.org/x/net/context"
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/rpc"
	"servlets/vcode"
	"utils"
	"utils/logger"
)

const (
	CHECK_TYPE_OF_PHONE = 1
	CHECK_TYPE_OF_EMAIL = 2

	ACCOUNT_EXISTS     = 1
	ACCOUNT_NOT_EXISTS = 2
)

type checkWithVcodeRequest struct {
	Base  *common.BaseInfo     `json:"base"`
	Param *checkWithVcodeParam `json:"param"`
}

type checkWithVcodeParam struct {
	ImgId    string `json:"img_id"`
	ImgVcode string `json:"img_vcode"`
	Type     int    `json:"type"`
	Country  int    `json:"country"`
	Phone    string `json:"phone"`
	EMail    string `json:"email"`
	Uid      string `json:"uid"`
}

type checkWithVcodeResponse struct {
	Exists int    `json:"exists"`
	Uid    string `json:"uid"`
	Status int64  `json:"status"`
}

type checkWithVcodeHandler struct {
}

func (handler *checkWithVcodeHandler) Method() string {
	return http.MethodPost
}

func (handler *checkWithVcodeHandler) Handle(request *http.Request, writer http.ResponseWriter) {
	response := &common.ResponseData{
		Base: &common.BaseResp{
			RC:  constants.RC_OK.Rc,
			Msg: constants.RC_OK.Msg,
		},
	}

	defer common.FlushJSONData2Client(response, writer)

	header := common.ParseHttpHeaderParams(request)
	data := checkWithVcodeRequest{}
	common.ParseHttpBodyParams(request, &data)

	if checkAccountRequestParams(header, &data) == false {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	if data.Param == nil {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	succFlag, code := vcode.ValidateImgVCode(data.Param.ImgId, data.Param.ImgVcode)
	//logger.Debug("succFlag:" + strconv.FormatBool(succFlag) + ",code:" + utils.Int2Str(code))
	if succFlag {
		resData := checkWithVcodeResponse{Exists: ACCOUNT_NOT_EXISTS}
		cli := rpc.GetUserCacheClient()
		if cli == nil {
			response.SetResponseBase(constants.RC_SYSTEM_ERR)
			return
		}
		switch data.Param.Type {
		case CHECK_TYPE_OF_EMAIL:
			req := &microuser.CheckAccountByEmailReq{
				Email: data.Param.EMail,
			}
			resp, err := cli.CheckAccountByEmail(context.Background(), req)
			if err != nil {
				response.SetResponseBase(constants.RC_SYSTEM_ERR)
				return
			}
			if resp.Result == microuser.ResCode_OK {
				resData.Exists = ACCOUNT_EXISTS
				resData.Uid = utils.Int642Str(resp.Uid)
				resData.Status = resp.Status
			}
		case CHECK_TYPE_OF_PHONE:
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
				resData.Exists = ACCOUNT_EXISTS
				resData.Uid = utils.Int642Str(resp.Uid)
				resData.Status = resp.Status
			}
		}
		response.Data = resData
	} else {
		response.SetResponseBase(vcode.ConvImgErr(code))
	}
}

func checkAccountRequestParams(header *common.HeaderParams, data *checkWithVcodeRequest) bool {
	if header.Timestamp < 1 {
		logger.Info("register user: no timestamp")
		return false
	}

	if (data.Base.App == nil) || (data.Base.App.IsValid() == false) {
		logger.Info("register user: app info is invalid")
		return false
	}

	if data.Param.Type != CHECK_TYPE_OF_EMAIL && data.Param.Type != CHECK_TYPE_OF_PHONE {
		logger.Info("check user: type invalid")
		return false
	}

	if data.Param.Type == CHECK_TYPE_OF_EMAIL && (utils.IsValidEmailAddr(data.Param.EMail) == false) {
		logger.Info("check account: email info invalid")
		return false
	}

	if data.Param.Type == CHECK_TYPE_OF_PHONE && (data.Param.Country == 0 || len(data.Param.Phone) < 1) {
		logger.Info("check user: phone info invalid")
		return false
	}

	return true
}
