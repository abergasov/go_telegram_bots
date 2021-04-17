package bot_list

import (
	"encoding/json"
	"errors"
	"go_telegram_bots/pkg/config"
	"go_telegram_bots/pkg/logger"

	tgbotapi "gopkg.in/go-telegram-bot-api/telegram-bot-api.v4"
)

type BelformagBot struct {
	abstractBot
}

func NewBelformagBot(conf *config.AppConfig, tWorker ITelegramWorker, buildTime, buildHash string) *BelformagBot {
	botName := "@belformag_bot"
	bf := &BelformagBot{}
	for i := range conf.BotList {
		if conf.BotList[i].BotName != botName {
			continue
		}
		bf.InitBot(tWorker, botName, &conf.BotList[i])
		break
	}
	if !bf.initialized {
		logger.Error("Belformag is not initializes", errors.New("config missing"))
	} else {
		bf.sendTelegramMessageToAdmin("Belformag started", buildHash, buildTime)
	}
	return bf
}

func (b *BelformagBot) HandleRequest(msg *tgbotapi.Update) {
	if msg.Message != nil {
		switch msg.Message.Text {
		case "/start", "/help":
			b.sendBotInfo(msg)
			return
		}
		if b.IsAdminChat(msg) {
			b.processCommand(msg)
		}
		b.LogEvent(msg)
	}
	if msg.CallbackQuery != nil {
		//o.handleCallback(msg.CallbackQuery.Data, msg.CallbackQuery.ID)
	}
}

func (b *BelformagBot) processCommand(msg *tgbotapi.Update) {

}

func (b *BelformagBot) sendBotInfo(msg *tgbotapi.Update) {
	b.botAPI.SendMessage(msg.Message.Chat.ID, "Hello world", nil, "")
}

func (b *BelformagBot) LogEvent(msg *tgbotapi.Update) {
	data, err := json.Marshal(msg)
	if err != nil {
		logger.Error("error log event", err)
		return
	}
	b.sendTelegramMessageToLog(string(data))
}
