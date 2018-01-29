package lvthttp

import (
	"encoding/json"
	"net/http"
	"net/url"
	"utils/logger"
	"net"
	"time"
	"io/ioutil"
)

var (
	client *HttpClien
)

func init() {
	client = NewDefaultHttpClient()
}

func dialTimeout(network, addr string) (net.Conn, error) {
	return net.DialTimeout(network, addr, time.Second*POST_REMOTE_TIMEOUT)
}

func NewHttpClient(keepAlives bool)*HttpClien{
	c := HttpClien{
		transport :&http.Transport{
			MaxIdleConns:1000,
			MaxIdleConnsPerHost:1000,
			Dial:              dialTimeout,
			DisableKeepAlives: !keepAlives, //为true时会在 body.Close()时关闭连接,不然close的时候也不会关闭链接
		},
	}
	c.build()
	return &c
}

func NewDefaultHttpClient()*HttpClien{
	return NewHttpClient(true)
}


func toJson(t interface{}) string {
	jsonByte, err := json.Marshal(t)
	if err != nil {
		logger.Error("json parse err ", err.Error())
		return ""
	}
	return string(jsonByte)
}

func map2UrlValues(p map[string]string) url.Values {
	if len(p) > 0 {
		uv := url.Values{}
		for i, v := range p {
			uv.Add(i, v)
		}
		return uv
	}
	return nil
}


func read(resp *http.Response) (string, error) {
	if resp == nil {
		return "",nil
	}
	body := resp.Body
	defer body.Close()
	res,err := ioutil.ReadAll(body)
	if err != nil {
		logger.Info("ParseHttpBodyParams: read http body error : ", err)
		return "", err
	}
	return string(res), nil
}

//发起post请求
func JsonPost(url string, params interface{}) (resBody string, e error) {
	return client.JsonPost(url,params)
}

func FormPost(url string, params map[string]string) (resBody string, e error) {
	return client.FormPost(url,params)
}

func Do(req *http.Request) (*http.Response, error) {
	return client.Do(req)
}
