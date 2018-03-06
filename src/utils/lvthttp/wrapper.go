package lvthttp

import (
	"net/http"
	"utils/logger"
	"strings"
	"net/url"
	"errors"
	"time"
)

const (
	POST_REMOTE_TIMEOUT = 30
)

type HttpClien struct{
	transport *http.Transport
	client *http.Client
	httpTimeout time.Duration
}


func (c *HttpClien)build(){
	c.client = &http.Client{
		Transport: c.transport,
		Timeout:c.httpTimeout,
	}
}



func (c *HttpClien)post(url, contentType, param string) (resp *http.Response, err error) {
	logger.Debug("SendPost url", url, "param", param)
	return c.client.Post(url, contentType, strings.NewReader(param))
}

func(c *HttpClien) postForm(url string, data url.Values) (resp *http.Response, err error) {
	logger.Debug("form http post url ", url, "param ", data)
	return c.client.PostForm(url, data)
}



//发起post请求
func (c *HttpClien)JsonPost(url string, params interface{}) (resBody string, e error) {
	jsonParam := toJson(params)
	resp, e1 := c.post(url, "application/json", jsonParam)
	if e1 != nil {
		logger.Error("send post error ---> ", e1.Error())
		return "", e1
	}
	return read(resp)
}

func (c *HttpClien)FormPost(url string, params map[string]string) (resBody string, e error) {
	resp, e1 := c.postForm(url, map2UrlValues(params))
	if e1 != nil {
		logger.Debug("post error ---> ", e1.Error())
		return "", e1
	}
	return read(resp)
}

func (c *HttpClien)Do(req *http.Request) (*http.Response, error) {
	res,err := c.client.Do(req)
	if err != nil {
		logger.Debug("post error ---> ", err.Error())
		return nil, err
	}
	if !checkHttpStatus(res.StatusCode) {
		return res, errors.New("http status "+res.Status)
	}
	return res,err
}