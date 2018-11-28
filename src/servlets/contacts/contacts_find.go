package contacts

import (
	"gitlab.maxthon.net/cloud/livesone-micro-user/src/proto"
	"golang.org/x/net/context"
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/rpc"
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
	cli := rpc.GetUserCacheClient()
	if cli == nil {
		res.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}
	var accs []int64 = make([]int64, 0)
	var err error
	if utils.IsValidEmailAddr(reqData.Param.Account) {
		req := &microuser.CheckAccountByEmailReq{
			Email:                reqData.Param.Account,
		}
		resp, rpcerr := cli.CheckAccountByEmail(context.Background(), req)
		if rpcerr != nil {
			err = rpcerr
		} else if resp.Result == microuser.ResCode_OK {
			accs = append(accs, resp.Uid)
		}
	} else {
		req := &microuser.CheckAccountByAccountReq{
			Account:reqData.Param.Account,
		}
		resp, rpcerr := cli.CheckAccountByAccount(context.Background(), req)
		if rpcerr != nil {
			err = rpcerr
		} else if resp.Result == microuser.ResCode_OK {
			accs = append(accs, resp.Uid)
		}
	}

	if err != nil {
		log.Error("query mysql db error", err.Error())
		res.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}

	res.Data = convContacts(accs)

}

func convContacts(accs []int64) []contactFindResData {
	if len(accs) > 0 {
		r := make([]contactFindResData, len(accs))
		for i, v := range accs {
			uid := v
			nickName ,_ := rpc.GetUserField(uid,microuser.UserField_NICKNAME)
			r[i] = contactFindResData{
				Uid:      uid,
				Nickname: nickName,
			}
		}
		return r
	}
	return nil
}
