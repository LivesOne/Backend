package server

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

var gRouter *gin.Engine
var gServer *http.Server

//Init http server object
func init() {
	gRouter = gin.Default()
}

//RegisterHandler
func RegisterHandler(url string, handler HttpHandler) {
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
		Addr:           ":8080",
		Handler:        gRouter,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	gServer.ListenAndServe()
}
