package contacts

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/token"
	"utils"
	"utils/logger"
)

type (
	contactDeleteHandler struct {

	}
	contactDeleteSecret struct {
		ContactId     int64        `json:"contact_id,omitempty"`
	}
	contactDeleteParam struct {
		Secret string `json:"secret"`
	}
	contactDeleteReqData struct {
		Param *contactDeleteParam `json:"param"`
	}
)

func (p *contactDeleteReqData) IsValid() bool {
	return p.Param != nil && len(p.Param.Secret) > 0
}

func (handler *contactDeleteHandler) Method() string {
	return http.MethodPost
}

func (handler *contactDeleteHandler) Handle(request *http.Request, writer http.ResponseWriter) {

	log := logger.NewLvtLogger(true, "contactDeleteHandler")
	defer log.InfoAll()

	res := common.NewResponseData()
	defer common.FlushJSONData2Client(res,writer)
	header := common.ParseHttpHeaderParams(request)
	if !header.IsValid() {
		log.Warn("header is not valid", utils.ToJSON(header))
		res.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	uidStr, aesKey, _, tokenErr := token.GetAll(header.TokenHash)
	if err := common.TokenErr2RcErr(tokenErr); err != constants.RC_OK {
		log.Info("get info from cache error:", err)
		res.SetResponseBase(err)
		return
	}

	reqData := new(contactDeleteReqData)

	if !common.ParseHttpBodyParams(request,reqData) {
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
	secret := new(contactDeleteSecret)
	iv, key := aesKey[:constants.AES_ivLen], aesKey[constants.AES_ivLen:]

	if err := utils.DecodeSecret(reqData.Param.Secret, key, iv, secret); err != nil {
		log.Info("decide secret failed")
		res.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	uid := utils.Str2Int64(uidStr)
	contactId := secret.ContactId

	if err := common.DeleteContact(uid, contactId); err != nil {
		log.Error("delete mongo failed", err.Error())
		res.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}
}