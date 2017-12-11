package server

import "net/http"

func StartServer() {
	if g_httpServer != nil {
		g_httpServer.start()
	}
}

func RegisterHandler(url string, handler HttpHandler) {
	if g_httpServer != nil {
		g_httpServer.registerHandler(url, handler)
	}
}

// func RegisterHandler(method int, handler HttpHandler) {
// 	if g_httpServer != nil {
// 		g_httpServer.registerHandler(method, handler)
// 	}
// }

type HttpHandler interface {
	Method() string
	Handle(*http.Request, http.ResponseWriter)
}
