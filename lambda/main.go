package main

import (
	"log"
	"os"

	"biclomap-be/lambda/ping"

	"github.com/apex/gateway"
	"github.com/gin-gonic/gin"
)

func routerEngine() *gin.Engine {
	gin.SetMode(gin.DebugMode)

	r := gin.New()

	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.GET("/ping", ping.Handler)
	return r
}

func main() {
	addr := ":" + os.Getenv("PORT")
	log.Fatal(gateway.ListenAndServe(addr, routerEngine()))
}
