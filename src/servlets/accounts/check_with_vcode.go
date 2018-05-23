package accounts

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"utils"
	"utils/logger"
	"utils/vcode"
)

const (
	CHECK_TYPE_OF_PHONE = 1
	CHECK_TYPE_OF_EMAIL = 2

	ACCOUNT_EXISTS     = 1
	ACCOUNT_NOT_EXISTS = 2
)

type checkWithVcodeRequest struct {
	Base  *common.BaseInfo            `json:"base"`
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
	Status int    `json:"status"`
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

		switch data.Param.Type {
		case CHECK_TYPE_OF_PHONE:
			uid, status := common.GetAssetByPhone(data.Param.Country, data.Param.Phone)
			if uid != 0 {
				resData.Exists = ACCOUNT_EXISTS
				resData.Status = status
				resData.Uid = utils.Int642Str(uid)
			}
		case CHECK_TYPE_OF_EMAIL:
			uid, status := common.GetAssetByEmail(data.Param.EMail)
			if uid != 0 {
				resData.Exists = ACCOUNT_EXISTS
				resData.Status = status
				resData.Uid = utils.Int642Str(uid)
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
