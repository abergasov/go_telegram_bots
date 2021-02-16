package bot_list

import (
	"encoding/json"
	"fmt"
	"go_telegram_bots/pkg/config"
	"go_telegram_bots/pkg/database"
	"go_telegram_bots/pkg/logger"
	"go_telegram_bots/pkg/utils"

	"go.uber.org/zap"

	tgbotapi "gopkg.in/go-telegram-bot-api/telegram-bot-api.v4"
)

type ITelegramWorker interface {
	SetToken(token string)
	DeleteMessage(chatID int64, messageID int)
	SendAlert(alert *utils.TelegramSendAlert)
	SendMessage(chat int64, message string, replyMarkup *utils.TelegramKeyboard, params ...string)
	SendMessagePrepared(tm *utils.TelegramSendMessage)
}

type abstractBot struct {
	hookURL     string
	token       string
	botName     string
	adminChats  []int64
	logChats    []int64
	initialized bool
	botAPI      ITelegramWorker
	db          *database.DBAdapter
}

func (a *abstractBot) InitBot(tW ITelegramWorker, botName string, botConf *config.BotConfig) {
	a.hookURL = botConf.BotHookPath
	a.botName = botName
	a.token = botConf.BotToken
	a.adminChats = botConf.BotAdminChats
	a.logChats = botConf.BotLoggerChat
	a.botAPI = tW
	a.initialized = true
	a.botAPI.SetToken(botConf.BotToken)
	if len(botConf.DBConf.DBHost) > 0 {
		a.initDB(&botConf.DBConf)
	}
}

func (a *abstractBot) initDB(conf *config.DBConf) {
	connect, err := database.InitConnection(conf)
	logger.Info("try to db", zap.String("host", conf.DBHost), zap.String("db", conf.DBName))
	if err != nil {
		logger.Fatal("Error db connect for bot "+a.botName, err)
	}
	logger.Info("connected to db", zap.String("host", conf.DBHost), zap.String("db", conf.DBName))
	a.db = connect
}

func (a *abstractBot) IsAdminChat(msg *tgbotapi.Update) bool {
	userID := 0
	if msg.Message != nil {
		userID = msg.Message.From.ID
	} else if msg.CallbackQuery != nil {
		userID = msg.CallbackQuery.From.ID
	}
	if userID == 0 {
		res, err := json.Marshal(msg)
		if err != nil {
			logger.Error("error marshal telegram message", err)
		} else {
			logger.Info("can't find userId from message", zap.String("msg", string(res)))
		}
	}
	for i := range a.adminChats {
		if a.adminChats[i] == int64(userID) {
			return true
		}
	}
	return false
}

func (a *abstractBot) GetMe() string {
	return a.botName
}

func (a *abstractBot) GetHookPath(hostUrl, hookPostfix string) (string, string) {
	hookStr := fmt.Sprintf("https://api.telegram.org/bot%s/setwebhook?url=%s%s%s", a.token, hostUrl, hookPostfix, a.hookURL)
	logger.Info(a.botName, zap.String("path", hookStr))
	return a.botName, hookPostfix + a.hookURL
}

func (a *abstractBot) sendTelegramMessageToAdmin(message string, params ...string) {
	for i := range a.adminChats {
		a.botAPI.SendMessage(a.adminChats[i], message, nil, params...)
	}
}

func (a *abstractBot) sendTelegramMessageToLog(message string, params ...string) {
	for i := range a.logChats {
		a.botAPI.SendMessage(a.logChats[i], message, nil, params...)
	}
}
