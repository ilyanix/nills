package main

import (
	"github.com/julienschmidt/httprouter"
)


func NewAPIServer() *httprouter.Router {
	router := httprouter.New()
	router.POST("/v1/join", Join)
	router.GET("/v1/nodes", ListNodes)
	return router
}
