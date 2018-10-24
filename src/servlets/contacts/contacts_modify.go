package contacts

import (
	"gopkg.in/mgo.v2"
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/token"
	"utils"
	"utils/logger"
)

type (
	contactModifyHandler struct {

	}
)


func (handler *contactModifyHandler) Method() string {
	return http.MethodPost
}

func (handler *contactModifyHandler) Handle(request *http.Request, writer http.ResponseWriter) {

	log := logger.NewLvtLogger(true, "contactListHandler")
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

	reqData := new(contactCreateReqData)
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
		res.SetResponseBase(constants.RC_INVALID_SIGN)
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

	mdfMap := convmap(secret)
	mdfMap["update_time"] = utils.GetTimestamp13()
	uid := utils.Str2Int64(uidStr)
	if err := common.ModifyContact(mdfMap,uid,secret.ContactId);err != nil || err != mgo.ErrNotFound {
		log.Error("update mongo  failed",err.Error())
		res.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}
}
