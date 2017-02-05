package goCMS_routes

import "github.com/gin-gonic/gin"

type ApiRoutes struct {
	Public *gin.RouterGroup
	PreTwofactor *gin.RouterGroup
	Auth   *gin.RouterGroup
	Admin *gin.RouterGroup
	Root *gin.RouterGroup
}
