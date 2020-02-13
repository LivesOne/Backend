package server

import "net/http"

type HttpHandler interface {
	Method() string
	Handle(*http.Request, http.ResponseWriter)
}
