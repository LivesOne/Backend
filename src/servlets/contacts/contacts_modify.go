package contacts

import (
	"github.com/gin-gonic/gin"
	"server"
	"servlets/common"
	"servlets/constants"
	"servlets/token"
	"utils"
	"utils/logger"
)

type (
	contactModifyHandler struct {
		server.DefHttpHandler
	}
)

func (vh *contactModifyHandler) Handle(c *gin.Context) {

	log := logger.NewLvtLogger(true, "contactCreateHandler")
	defer log.InfoAll()

	res := common.NewResponseData()
	defer vh.RJson(c, res)
	header := vh.GetHeadParams(c)
	if !header.IsValid() {
		log.Warn("header is not valid", utils.ToJSON(header))
		res.SetResponseBase(constants.CODE_PARAM_ERR)
		return
	}

	uidStr, aesKey, _, tokenErr := token.GetAll(header.TokenHash)
	if err := common.TokenErr2RcErr(tokenErr); err != constants.CODE_OK {
		log.Info("get info from cache error:", err)
		res.SetResponseBase(err)
		return
	}

	reqData := new(contactCreateReqData)
	if err := c.ShouldBindJSON(reqData); err != nil {
		log.Info("decode json str error", err)
		res.SetResponseBase(constants.CODE_PROTOCOL_ERR)
		return
	}
	if !reqData.IsValid() {
		log.Info("required param is nil")
		res.SetResponseBase(constants.CODE_PARAM_ERR)
		return
	}

	if len(aesKey) != constants.AES_totalLen {
		log.Info(" get aeskey from cache error:", len(aesKey))
		res.SetResponseBase(constants.CODE_SYSTEM_ERR)
		return
	}

	if !utils.SignValid(aesKey, header.Signature, header.Timestamp) {
		log.Info("validate sign failed ")
		res.SetResponseBase(constants.CODE_INVALID_SIGN)
		return
	}

	// 解码 secret 参数
	// secretString := requestData.Param.Secret
	secret := new(contactCreateSecret)
	iv, key := aesKey[:constants.AES_ivLen], aesKey[constants.AES_ivLen:]

	if err := utils.DecodeSecret(reqData.Param.Secret, key, iv, secret); err != nil {
		log.Info("decide secret failed")
		res.SetResponseBase(constants.CODE_PARAM_ERR)
		return
	}

	mdfMap := convmap(secret)
	uid := utils.Str2Int64(uidStr)
	if err := common.ModifyContact(mdfMap,uid,secret.ContactId);err != nil {
		log.Error("insert mongo  failed",err.Error())
		res.SetResponseBase(constants.CODE_SYSTEM_ERR)
		return
	}
}
