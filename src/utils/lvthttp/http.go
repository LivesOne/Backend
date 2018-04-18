package lvthttp

import (
	"utils/logger"
	"github.com/levigross/grequests"
	"strings"
	"errors"
	"utils"
)


//发起post请求
func JsonPost(url string, params interface{}) (string, error) {
	ro := &grequests.RequestOptions{
		JSON: params,
	}
	logger.Info("http post url [",url,"] param :",utils.ToJSON(params))
	resp, err := grequests.Post(url,ro)
	if err != nil {
		logger.Error("http error",err.Error())
		return "",err
	}
	return getRes(resp)
}

func FormPost(url string, params map[string]string) (string, error) {
	ro := &grequests.RequestOptions{
		Params: params,
	}
	resp, err := grequests.Post(url,ro)
	if err != nil {
		return "",err
	}
	return getRes(resp)
}

func Get(url string,params map[string]string) (string, error) {
	ro := &grequests.RequestOptions{
		Params: params,
	}
	resp, err := grequests.Get(url,ro)
	if err != nil {
		return "",err
	}
	return getRes(resp)
}

func getRes(resp *grequests.Response)(string,error){
	if resp.Ok {
		r := strings.TrimSpace(resp.String())
		return r,nil
	}
	return "",errors.New("http req faild")
}