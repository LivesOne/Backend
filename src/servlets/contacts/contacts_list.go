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
	contactListHandler struct {

	}
)


func (handler *contactListHandler) Method() string {
	return http.MethodPost
}

func (handler *contactListHandler) Handle(request *http.Request, writer http.ResponseWriter) {

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

	uidStr, _, _, tokenErr := token.GetAll(header.TokenHash)
	if err := common.TokenErr2RcErr(tokenErr); err != constants.RC_OK {
		log.Info("get info from cache error:", err)
		res.SetResponseBase(err)
		return
	}
	uid := utils.Str2Int64(uidStr)

	contactList := common.GetContactsListByUid(uid)
	if contactList == nil {
		log.Error("query mongo error")
		res.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}
	if len(contactList) > 0 {
		res.Data = contactList
	}

}
