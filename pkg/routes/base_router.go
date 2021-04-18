package routes

import (
	"net/http"

	"github.com/abergasov/go_telegram_bots/pkg/bot_manager"
	"github.com/abergasov/go_telegram_bots/pkg/config"
	tgbotapi "gopkg.in/go-telegram-bot-api/telegram-bot-api.v4"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type AppRouter struct {
	GinEngine  *gin.Engine
	BotManager *bot_manager.BotManager
	config     *config.AppConfig
	upgrader   websocket.Upgrader
}

func InitRouter(cnf *config.AppConfig, hookPostfix, buildTime, buildHash string) *AppRouter {
	if cnf.ProdEnv {
		gin.SetMode(gin.ReleaseMode)
	}
	return &AppRouter{
		GinEngine:  gin.Default(),
		BotManager: bot_manager.InitBotManager(cnf, hookPostfix, buildTime, buildHash),
		config:     cnf,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
			HandshakeTimeout: 0,
			//ReadBufferSize:  1024,
			//WriteBufferSize: 1024,
		},
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
