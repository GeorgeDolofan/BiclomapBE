package main

import (
	"log"
	"os"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"biclomap-be/lambda/awscontext"
	"biclomap-be/lambda/facebook"
	"biclomap-be/lambda/login"
	"biclomap-be/lambda/ping"

	"github.com/apex/gateway"
	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func routerEngine() *gin.Engine {
	gin.SetMode(gin.DebugMode)

	r := gin.New()

	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(CORSMiddleware())
	r.Use(awscontext.AWSContext())

	swagger_url := ginSwagger.URL("https://dev.biclomap.com/swagger/doc.json")
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, swagger_url))

	r.GET("/ping", ping.Handler)

	r.POST("/login/facebook", login.Facebook)
	r.POST("/login/email", login.Email)
	r.POST("/login/email/signup", login.EmailSignup)

	r.GET("/fb/redirect", facebook.Redirect)
	r.GET("/fb/deauthorize", facebook.Deauthorize)
	return r
}

// @title Biclomap REST API
// @version 1
// @description This is the Biclomap back-end server
func main() {
	addr := ":" + os.Getenv("PORT")
	log.Fatal(gateway.ListenAndServe(addr, routerEngine()))
}
