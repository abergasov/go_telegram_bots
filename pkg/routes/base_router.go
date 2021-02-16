package routes

import (
	"encoding/json"
	"errors"
	"go.uber.org/zap"
	"go_telegram_bots/pkg/bot_manager"
	"go_telegram_bots/pkg/bot_manager/bot_list"
	"go_telegram_bots/pkg/config"
	"go_telegram_bots/pkg/logger"
	tgbotapi "gopkg.in/go-telegram-bot-api/telegram-bot-api.v4"
	"net/http"
	"time"

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

func (a *AppRouter) HandleClientRest(c *gin.Context) {
	controllerBot := a.BotManager.GetBot("@rmcpi_bot")
	if controllerBot == nil {
		logger.Error("controller bot not registred", errors.New("can't find bot"))
		return
	}
	bot, ok := controllerBot.(*bot_list.ControllerBot)
	if !ok {
		logger.Error("can't convert SingleBot to ControllerBot", errors.New("can't find bot"))
	}
	var cmd *bot_list.Command
	bot.MuCommand.Lock()
	if bot.ActiveCommand != nil {
		cmd = bot.ActiveCommand
		bot.ActiveCommand = nil
	}
	bot.MuCommand.Unlock()
	c.JSON(http.StatusOK, gin.H{"ok": true, "cmd": cmd})
}

func (a *AppRouter) HandleClient(w http.ResponseWriter, r *http.Request, ip string) {
	c, err := a.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("Upgrade connection error", err)
		return
	}
	defer c.Close()
	controllerBot := a.BotManager.GetBot("@rmcpi_bot")
	if controllerBot == nil {
		logger.Error("controller bot not registred", errors.New("can't find bot"))
		return
	}

	bot, ok := controllerBot.(*bot_list.ControllerBot)
	if !ok {
		logger.Error("can't convert SingleBot to ControllerBot", errors.New("can't find bot"))
	}

	var cmd *bot_list.Command
	for {
		bot.MuCommand.Lock()
		if bot.ActiveCommand != nil {
			cmd = bot.ActiveCommand
			bot.ActiveCommand = nil
		}
		bot.MuCommand.Unlock()

		if cmd != nil {
			b, _ := json.Marshal(cmd)
			err = c.WriteMessage(websocket.TextMessage, b)
			if err != nil {
				logger.Error("Error write in socket", err)
				break
			}
			logger.Info("received message", zap.String("ip", ip), zap.String("data", string(b)))
		}
		cmd = nil
		time.Sleep(300 * time.Millisecond)
	}
	logger.Info("Finish write in connection")
}
