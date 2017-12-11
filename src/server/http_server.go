package server

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

//define the http server struct
type httpServerStruct struct {
	postHandlers map[string]HttpHandler
	getHandlers  map[string]HttpHandler
}

// package level "global" data
var g_httpServer *httpServerStruct

// initialize function
func init() {
	g_httpServer = new(httpServerStruct)
	g_httpServer.init()
}

// initialize http server object
func (server *httpServerStruct) init() {
	server.postHandlers = make(map[string]HttpHandler)
	server.getHandlers = make(map[string]HttpHandler)
}

func (server *httpServerStruct) registerHandler(url string, handler HttpHandler) {

	// r := gin.Default()
	var handlersMap map[string]HttpHandler

	switch handler.Method() {
	case http.MethodGet:
		// fmt.Println("------------- registerHandler", handler.Method, url)
		// r.GET(url, server.createPostHandler_JSON_API(handler))
		handlersMap = server.getHandlers
	case http.MethodPost:
		// fmt.Println("------------- registerHandler", handler.Method, url)
		// r.POST(url, server.createPostHandler_JSON_API(handler))
		handlersMap = server.postHandlers
	}

	if _, present := handlersMap[url]; present {
		return
	}
	handlersMap[url] = handler

	// r.Run()

}

// start http server to listen
func (server *httpServerStruct) start() {

	r := gin.Default()

	for url, handler := range server.postHandlers {
		log.Println("[Info] register http POST handler=" + url)
		r.POST(url, server.createPostHandler_JSON_API(handler))
	}

	for url, handler := range server.getHandlers {
		log.Println("[Info] register http GET handler=" + url)
		r.GET(url, server.createGetHandler_JSON_API(handler))
	}

	server.postHandlers = nil
	server.getHandlers = nil

	r.Run()
}

func (server *httpServerStruct) createGetHandler_JSON_API(handler HttpHandler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// server.handleGetAction(handler, ctx)
		handler.Handle(ctx.Request, ctx.Writer)
	}
}

// func (server *httpServerStruct) handleGetAction(handler HttpHandler, ctx *gin.Context) {
// 	handler.Handle(ctx.Request, ctx.Writer)
// }

func (server *httpServerStruct) createPostHandler_JSON_API(handler HttpHandler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// server.handlePostAction(handler, ctx)
		handler.Handle(ctx.Request, ctx.Writer)
	}
}

// func (server *httpServerStruct) handlePostAction(handler HttpHandler, ctx *gin.Context) {

// 	// fmt.Println("handlePostAction 1234:", ctx.PostForm("param"))
// 	handler.Handle(ctx.Request, ctx.Writer)
// }
