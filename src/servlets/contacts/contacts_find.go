package contacts

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"utils"
	"utils/logger"
)

type (
	contactFindHandler struct {
	}
	contactFindParam struct {
		Account string `json:"account"`
	}
	contactFindReqData struct {
		Base  *common.BaseInfo  `json:"base"`
		Param *contactFindParam `json:"param"`
	}

	contactFindResData struct {
		Uid      int64  `json:"uid"`
		Nickname string `json:"nickname"`
	}
)

func (p *contactFindReqData) IsValid() bool {
	return p.Param != nil && len(p.Param.Account) > 0
}

func (handler *contactFindHandler) Method() string {
	return http.MethodPost
}

func (handler *contactFindHandler) Handle(request *http.Request, writer http.ResponseWriter) {

	log := logger.NewLvtLogger(true, "contactFindHandler")
	defer log.InfoAll()

	res := common.NewResponseData()
	defer common.FlushJSONData2Client(res, writer)

	reqData := new(contactFindReqData)

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

	// 解码 secret 参数
	// secretString := requestData.Param.Secret

	var accs []*common.Account = make([]*common.Account, 0)
	var err error
	if utils.IsValidEmailAddr(reqData.Param.Account) {
		acc, errr := common.GetAccountByEmail(reqData.Param.Account)
		if errr != nil {
			err = errr
		} else if acc != nil {
			accs = append(accs, acc)
		}
	} else {
		accs, err = common.GetAccountListByPhoneOrUID(reqData.Param.Account)
	}

	if err != nil {
		log.Error("query mysql db error", err.Error())
		res.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}

	res.Data = convContacts(accs)

}

func convContacts(accs []*common.Account) []contactFindResData {
	if len(accs) > 0 {
		r := make([]contactFindResData, len(accs))
		for i, v := range accs {
			r[i] = contactFindResData{
				Uid:      v.UID,
				Nickname: v.Nickname,
			}
		}
		return r
	}
	return nil
}
