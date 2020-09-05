package server

import (
	"github.com/gin-gonic/gin"
	"github.com/lng50k/booster-backend/controllers"
	"github.com/lng50k/booster-backend/middlewares"
)

func NewRouter() *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	health := new(controllers.HealthController)

	router.GET("/health", health.Status)
	// router.Use(middlewares.AuthMiddleware())
	router.Use(middlewares.CORSMiddleware())

	v1 := router.Group("api/v1")
	{
		userGroup := v1.Group("user")
		{
			user := new(controllers.UserController)
			userGroup.GET("/:id", user.Retrieve)
		}

		whmGroup := v1.Group("whm")
		{
			whm := new(controllers.WHMController)
			whmGroup.GET("listaccounts", whm.RetrieveAll)
		}
	}
	return router

}
