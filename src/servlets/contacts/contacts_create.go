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
	contactCreateHandler struct {
	}
	extendInfo struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	contactCreateSecret struct {
		ContactId     int64        `json:"contact_id,omitempty"`
		Email         string       `json:"email,omitempty"`
		Name          string       `json:"name,omitempty"`
		Remark        string       `json:"remark,omitempty"`
		Country       int          `json:"country,omitempty"`
		Phone         string       `json:"phone,omitempty"`
		LivesoneUid   string       `json:"livesone_uid,omitempty"`
		WalletAddress string       `json:"wallet_address,omitempty"`
		Extend        []extendInfo `json:"extend,omitempty"`
	}
	contactCreateParam struct {
		Secret string `json:"secret"`
	}
	contactCreateReqData struct {
		Param *contactCreateParam `json:"param"`
	}
)

func (p *contactCreateReqData) IsValid() bool {
	return p.Param != nil && len(p.Param.Secret) > 0
}

func (handler *contactCreateHandler) Method() string {
	return http.MethodPost
}

func (handler *contactCreateHandler) Handle(request *http.Request, writer http.ResponseWriter) {

	log := logger.NewLvtLogger(true, "contactCreateHandler")
	defer log.InfoAll()

	res := common.NewResponseData()
	defer common.FlushJSONData2Client(res, writer)
	header := common.ParseHttpHeaderParams(request)
	if !header.IsValid() {
		log.Warn("header is not valid", utils.ToJSON(header))
		res.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	uidStr, aesKey, _, tokenErr := rpc.GetTokenInfo(header.TokenHash)
	if err := rpc.TokenErr2RcErr(tokenErr); err != constants.RC_OK {
		log.Info("get info from cache error:", err)
		res.SetResponseBase(err)
		return
	}

	reqData := new(contactCreateReqData)

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

	// 解码 secret 参数
	// secretString := requestData.Param.Secret
	secret := new(contactCreateSecret)
	iv, key := aesKey[:constants.AES_ivLen], aesKey[constants.AES_ivLen:]

	if err := utils.DecodeSecret(reqData.Param.Secret, key, iv, secret); err != nil {
		log.Info("decide secret failed")
		res.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	insertMap := convmap(secret)
	insertMap["uid"] = utils.Str2Int64(uidStr)
	insertMap["create_time"] = utils.GetTimestamp13()
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
}

func convmap(s *contactCreateSecret) map[string]interface{} {
	m := utils.StructConvMap(s)
	for _, v := range s.Extend {
		m[v.Key] = v.Value
	}
	return m
}
