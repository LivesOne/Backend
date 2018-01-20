package lvthttp

import (
	"net/http"
	"utils/logger"
	"strings"
	"io/ioutil"
	"net"
	"time"
	"encoding/json"
	"net/url"
)

const(
	POST_REMOTE_TIMEOUT = 30
)

var (
	client http.Client
)

func dialTimeout(network, addr string) (net.Conn, error) {
	return net.DialTimeout(network, addr, time.Second*POST_REMOTE_TIMEOUT)
}

func init(){
	transport := http.Transport{
		Dial:              dialTimeout,
		DisableKeepAlives: true,//为true时会在 body.Close()时关闭连接,不然close的时候也不会关闭链接
	}
	client = http.Client{
		Transport: &transport,
	}
}

func toJson(t interface{}) string {
	jsonByte, err := json.Marshal(t)
	if err != nil {
		logger.Error("json parse err ", err.Error())
		return ""
	}
	return string(jsonByte)
}

func map2UrlValues(p map[string]string)url.Values{
	if len(p) > 0 {
		uv := make(url.Values,0)
		for i,v := range p{
			uv[i] = []string{v}
		}
		return uv
	}
	return nil
}

func post(url,contentType,param string)(resp *http.Response, err error){
	return client.Post(url, contentType, strings.NewReader(param))
}

func postForm(url string,data url.Values)(resp *http.Response, err error){
	return client.PostForm(url, data)
}

//发起post请求
func JsonPost(url string, params interface{}) (resBody string, e error) {
	jsonParam := toJson(params)
	logger.Info("SendPost url", url, "param", jsonParam)
	resp, e1 := post(url, "application/json", jsonParam)
	if e1 != nil {
		logger.Error("send post error ---> ", e1.Error())
		return "", e1
	} else {
		defer resp.Body.Close()
		body, e2 := ioutil.ReadAll(resp.Body)
		if e2 != nil {
			logger.Error("post read error ---> ", e2.Error())
		}
		resBody := string(body)
		logger.Info("http res", resBody)
		return resBody, e2
	}
}

func FormPost(url string,params map[string]string)(resBody string, e error){
	logger.Info("form http post url ",url,"param ",params)
	resp, e1 := postForm(url,map2UrlValues(params))
	if e1 != nil {
		logger.Error("post error ---> ", e1.Error())
		return "", e1
	} else {
		defer resp.Body.Close()
		body, e2 := ioutil.ReadAll(resp.Body)
		if e2 != nil {
			logger.Error("post error ---> ", e2.Error())
		}
		res := string(body)
		logger.Info("SendPost res ---> ", res)
		return res, e2
	}
}
