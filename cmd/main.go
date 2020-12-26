package main

import (
	"go_bots/pkg/config"
	"go_bots/pkg/logger"
	"go_bots/pkg/middleware"
	"go_bots/pkg/routes"
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	hookPostfix = "/bot/"
	buildTime   = "_dev"
	buildHash   = "_dev"
)

func main() {
	logger.NewLogger()
	appConf := config.InitConf("/configs/conf.yaml")
	router := routes.InitRouter(appConf, hookPostfix, buildTime, buildHash)
	for i := range router.BotManager.HooksURLList {
		// для каждого бота регаем свой обработчик хуков, исходя из того постфикса, что определен у бота
		router.GinEngine.POST(i, router.HandleHook)
	}
	router.GinEngine.GET("/bot/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"ok":         true,
			"build_time": buildTime,
			"build_hash": buildHash,
		})
	})

	authorized := router.GinEngine.Group("/bot/ws/client/")
	authorized.Use(middleware.AuthControllerMiddleware(appConf.KeyToken))
	authorized.GET("pi", func(c *gin.Context) {
		router.HandleClient(c.Writer, c.Request)
	})

	logger.Info("Starting server at port " + appConf.AppPort)

	errR := router.GinEngine.Run(":" + appConf.AppPort)
	if errR != nil {
		logger.Fatal("Common server error", errR)
	}
}
