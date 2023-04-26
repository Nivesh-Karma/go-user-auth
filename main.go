package main

import (
	"github.com/Nivesh-Karma/go-user-admin/config"
	"github.com/Nivesh-Karma/go-user-admin/routes"
	"github.com/gin-gonic/gin"
)

func init() {
	config.Loadenv()
	config.ConnectDB()
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	routes.Route(router)
	router.Run(":4050") // listen and serve on 0.0.0.0:4050
}
