package main

import (
	"github.com/julienschmidt/httprouter"
)


func NewAPIServer() *httprouter.Router {
	router := httprouter.New()
	router.POST("/v1/join", handlJoin)
	router.GET("/v1/nodes", handlListNodes)
	return router
}
