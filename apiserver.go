package main

import (
	"github.com/julienschmidt/httprouter"
)


func NewAPIServer() *httprouter.Router {
	router := httprouter.New()
	router.GET("/v1/join", Join)
	Trace.Println("API routes created")
	return router
}
