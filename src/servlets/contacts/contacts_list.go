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
	contactListHandler struct {
		server.DefHttpHandler
	}
)

func (vh *contactListHandler) Handle(c *gin.Context) {

	log := logger.NewLvtLogger(true, "contactListHandler")
	defer log.InfoAll()

	res := common.NewResponseData()
	defer vh.RJson(c, res)
	header := vh.GetHeadParams(c)
	if !header.IsValid() {
		log.Warn("header is not valid", utils.ToJSON(header))
		res.SetResponseBase(constants.CODE_PARAM_ERR)
		return
	}

	uidStr, _, _, tokenErr := token.GetAll(header.TokenHash)
	if err := common.TokenErr2RcErr(tokenErr); err != constants.CODE_OK {
		log.Info("get info from cache error:", err)
		res.SetResponseBase(err)
		return
	}
	uid := utils.Str2Int64(uidStr)

	contactList := common.GetContactsListByUid(uid)
	if contactList == nil {
		log.Error("query mongo error")
		res.SetResponseBase(constants.CODE_SYSTEM_ERR)
		return
	}
	if len(contactList) == 0 {
		res.Data = contactList
	}

}
