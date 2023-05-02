package routes

import (
	"github.com/Nivesh-Karma/go-user-admin/controller"
	"github.com/Nivesh-Karma/go-user-admin/middleware"
	"github.com/gin-gonic/gin"
)

func Route(router *gin.Engine) {
	router.Use(CORSMiddleware())
	v1 := router.Group("/api/v1")
	{
		v1.POST("/create", controller.CreateNewUser)
		v1.POST("/login", controller.Login)
		v1.PATCH("/reset-password", controller.ResetPassword)
		v1.PATCH("/unlock-account", controller.UnlockAccount)
		v1.GET("/verify-token", middleware.RequireAuth, controller.Validate)
		v1.GET("/verify-admin", middleware.RequireAuth, controller.ValidateAdmin)
		v1.GET("/admin-updates", middleware.RequireAuth, controller.AdminUpdates)
		v1.POST("/refresh-token", middleware.VerifyRefreshToken, controller.RefreshUserToken)
	}

	google := router.Group("/google-auth/api/v1")
	{
		google.POST("/token", controller.GoogleLogin)
	}
	default_route := router.Group("/")
	{
		default_route.GET("healthcheck", func(ctx *gin.Context) {
			ctx.JSON(200, gin.H{
				"message": "success",
			})
		})
	}

}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "false")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST,HEAD,PATCH,OPTIONS,GET,PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
