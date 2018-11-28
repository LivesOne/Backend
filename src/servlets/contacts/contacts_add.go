package contacts

import (
	"gopkg.in/mgo.v2"
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/rpc"
	"utils"
	"utils/logger"
)

type (
	contactAddHandler struct {
	}
	contactAddResData struct {
		ContactId     int64  `json:"contact_id,omitempty"`
		Name          string `json:"name,omitempty"`
		Remark        string `json:"remark,omitempty"`
		Email         string `json:"email,omitempty"`
		Country       int    `json:"country,omitempty"`
		Phone         string `json:"phone,omitempty"`
		LivesoneUid   string `json:"livesone_uid,omitempty"`
		WalletAddress string `json:"wallet_address,omitempty"`
		CreateTime    int64  `json:"create_time,omitempty"`
		Uid           int64  `json:"-"`
	}
	contactAddParam struct {
		Uid       string `json:"uid"`
		ContactId int64  `json:"contact_id"`
	}
	contactAddReqData struct {
		Param *contactAddParam `json:"param"`
	}
)

func (p *contactAddReqData) IsValid() bool {
	return p.Param != nil && len(p.Param.Uid) > 0 && p.Param.ContactId > 0
}

func (handler *contactAddHandler) Method() string {
	return http.MethodPost
}

func (handler *contactAddHandler) Handle(request *http.Request, writer http.ResponseWriter) {

	log := logger.NewLvtLogger(true, "contactAddHandler")
	defer log.InfoAll()

	res := common.NewResponseData()
	defer common.FlushJSONData2Client(res, writer)
	header := common.ParseHttpHeaderParams(request)
	if !header.IsValid() {
		log.Warn("header is not valid", utils.ToJSON(header))
		res.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	// 判断用户身份
	uidStr, aesKey, _, tokenErr := rpc.GetTokenInfo(header.TokenHash)
	if err := rpc.TokenErr2RcErr(tokenErr); err != constants.RC_OK {
		logger.Info("asset lockList: get info from cache error:", err)
		res.SetResponseBase(err)
		return
	}


	reqData := new(contactAddReqData)

	if !common.ParseHttpBodyParams(request, reqData) {
		log.Info("decode json str error")
		res.SetResponseBase(constants.RC_PROTOCOL_ERR)
		return
	}
	if !reqData.IsValid() {
		log.Info("required param is nil")
		res.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	if len(aesKey) != constants.AES_totalLen {
		log.Info(" get aeskey from cache error:", len(aesKey))
		res.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	if !utils.SignValid(aesKey, header.Signature, header.Timestamp) {
		log.Info("validate sign failed ")
		res.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	uid := utils.Str2Int64(uidStr)
	tagUid := utils.Str2Int64(reqData.Param.Uid)

	acc, err := rpc.GetUserInfo(tagUid)
	if err != nil {
		res.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}
	if acc == nil {
		res.SetResponseBase(constants.RC_INVALID_ACCOUNT)
		return
	}
	data := &contactAddResData{
		ContactId:     reqData.Param.ContactId,
		Name:          acc.Nickname,
		Remark:        "",
		Email:         acc.Email,
		Country:       int(acc.Country),
		LivesoneUid:   reqData.Param.Uid,
		Phone:         acc.Phone,
		WalletAddress: acc.WalletAddress,
		CreateTime:    utils.GetTimestamp13(),
		Uid:           uid,
	}

	insertMap := utils.StructConvMap(data)
	insertMap["livesone_uid"] = tagUid
	if err := common.CreateContact(insertMap); err != nil {
		log.Error("insert mongo  failed", err.Error())
		if mgo.IsDup(err) {
			res.SetResponseBase(constants.RC_DUP_CONTACT_ID)
			return
		}

		if err == mgo.ErrNotFound {
			res.SetResponseBase(constants.RC_CONTACT_ID_NOT_EXISTS)
			return
		}
		res.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}

	res.Data = data
}
