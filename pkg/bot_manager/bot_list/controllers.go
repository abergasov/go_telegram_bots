package bot_list

import (
	"encoding/json"
	"errors"
	"go_bots/pkg/config"
	"go_bots/pkg/logger"
	tgbotapi "gopkg.in/go-telegram-bot-api/telegram-bot-api.v4"
)

type ControllerBot struct {
	abstractBot
}

func NewControllerBot(conf *config.AppConfig, tWorker ITelegramWorker, buildTime, buildHash string) *ControllerBot {
	botName := "@rmcpi_bot"
	mB := &ControllerBot{}
	for i := range conf.BotList {
		if conf.BotList[i].BotName != botName {
			continue
		}
		mB.InitBot(tWorker, botName, &conf.BotList[i])
		break
	}
	if !mB.initialized {
		logger.Error("ControllerBot is not initializes", errors.New("config missing"))
	} else {
		mB.sendTelegramMessageToAdmin("ControllerBot started", buildHash, buildTime)
	}
	return mB
}

func (o *ControllerBot) HandleRequest(msg *tgbotapi.Update) {
	if msg.Message != nil {
		switch msg.Message.Text {
		case "/start", "/help":
			o.sendBotInfo(msg)
			return
		}
	}
}

func (o *ControllerBot) LogEvent(msg *tgbotapi.Update) {
	data, err := json.Marshal(msg)
	if err != nil {
		logger.Error("error log event", err)
		return
	}
	o.sendTelegramMessageToLog(string(data))
}

func (o *ControllerBot) sendBotInfo(msg *tgbotapi.Update) {
	o.botAPI.SendMessage(msg.Message.Chat.ID, "Bot allow control pi", nil)
}
