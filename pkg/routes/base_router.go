package routes

import (
	"go_bots/pkg/bot_manager"
	"go_bots/pkg/config"
	"net/http"

	tgbotapi "gopkg.in/go-telegram-bot-api/telegram-bot-api.v4"

	"github.com/gin-gonic/gin"
)

type AppRouter struct {
	GinEngine  *gin.Engine
	BotManager *bot_manager.BotManager
	config     *config.AppConfig
}

func InitRouter(cnf *config.AppConfig, hookPostfix, buildTime, buildHash string) *AppRouter {
	if cnf.ProdEnv {
		gin.SetMode(gin.ReleaseMode)
	}
	return &AppRouter{
		GinEngine:  gin.Default(),
		BotManager: bot_manager.InitBotManager(cnf, hookPostfix, buildTime, buildHash),
		config:     cnf,
	}
}

func (a *AppRouter) HandleHook(c *gin.Context) {
	var m tgbotapi.Update
	if err := c.ShouldBindJSON(&m); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "Invalid message"})
		return
	}
	for i := range a.BotManager.HooksURLList {
		if c.FullPath() != i {
			continue
		}
		a.BotManager.HandleTelegramRequest(&m, a.BotManager.HooksURLList[i])
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"ok": false})
}
