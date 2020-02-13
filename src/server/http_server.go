package server

import (
	"net/http"
	"time"

	"fmt"
	"github.com/gin-gonic/gin"
	"reflect"
	"runtime/debug"
	"servlets/common"
	"servlets/constants"
	"utils/logger"
)

var gRouter *gin.Engine
var gServer *http.Server

//Init http server object
func init() {
	gin.SetMode(gin.ReleaseMode)
	gRouter = gin.Default()
	gRouter.Use(globalRecover)
	gRouter.Use(logUrl)
}

//RegisterHandler
func RegisterHandler(url string, handler HttpHandler) {
	registerHandlerLog(url, handler)
	switch handler.Method() {
	case http.MethodGet:
		gRouter.GET(url, func(ctx *gin.Context) {
			handler.Handle(ctx.Request, ctx.Writer)
		})
		break
	case http.MethodPost:
		gRouter.POST(url, func(ctx *gin.Context) {
			handler.Handle(ctx.Request, ctx.Writer)
		})
		break
	default:
		break
	}
}

//Start http server to listen
func Start(addr string) {
	gServer = &http.Server{
		Addr:           addr,
		Handler:        gRouter,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	gServer.ListenAndServe()
}

func globalRecover(c *gin.Context) {
	defer func(c *gin.Context) {
		if rec := recover(); rec != nil {
			logger.Error("server panic: ", rec, string(debug.Stack()))
			response := common.NewResponseData()
			response.SetResponseBase(constants.RC_SYSTEM_ERR)
			c.JSON(http.StatusOK, response)
		}
	}(c)
	c.Next()
}

func logUrl(c *gin.Context) {
	req := c.Request
	logger.Info("recive req method 【", req.Method, "】 url ---> ", req.URL)
	c.Next()
}

func registerHandlerLog(url string, handler HttpHandler) {
	t := reflect.ValueOf(handler).Type()
	s := fmt.Sprintf("register req method %s url %s ---> %s", handler.Method(), url, t.String())
	logger.Info(s)
}
