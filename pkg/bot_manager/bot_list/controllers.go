package bot_list

import (
	"encoding/json"
	"errors"
	"go.uber.org/zap"
	"go_bots/pkg/config"
	"go_bots/pkg/logger"
	"go_bots/pkg/utils"
	tgbotapi "gopkg.in/go-telegram-bot-api/telegram-bot-api.v4"
	"net/url"
	"strings"
	"sync"
)

const (
	SPEED_0_75   = "a"
	SPEED_1      = "b"
	SPEED_1_25   = "c"
	SPEED_1_5    = "d"
	SPEED_1_75   = "e"
	SPEED_2      = "f"
	FORWARD      = "g"
	FORWARD_MORE = "h"
	BACK         = "i"
	BACK_MORE    = "j"
	PLAY         = "k"
	CLOSE        = "l"
)

type Command struct {
	Cmd      string // команда для выполнения
	ActionID string // для какого видео
}

type ControllerBot struct {
	abstractBot
	MuCommand     *sync.RWMutex
	ActiveCommand *Command
}

func NewControllerBot(conf *config.AppConfig, tWorker ITelegramWorker, buildTime, buildHash string) *ControllerBot {
	botName := "@rmcpi_bot"
	mB := &ControllerBot{
		MuCommand: &sync.RWMutex{},
	}
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
		if o.IsAdminChat(msg) {
			o.processYouTube(msg)
		}
	}
	if msg.CallbackQuery != nil {
		o.handleCallback(msg.CallbackQuery.Data, msg.CallbackQuery.ID)
	}
}

func (o *ControllerBot) handleCallback(callbackMessage, callbackQueryID string) {
	msg := strings.ReplaceAll(callbackMessage, "/", "")
	data := strings.Split(msg, "_")
	o.MuCommand.Lock()
	o.ActiveCommand = &Command{
		Cmd:      data[0],
		ActionID: data[1],
	}
	o.MuCommand.Unlock()
	logger.Info("message to control", zap.String("command", callbackMessage))
	o.botAPI.SendAlert(&utils.TelegramSendAlert{
		Text:            "added",
		ShowAlert:       false,
		CallbackQueryID: callbackQueryID,
	})
}

func (o *ControllerBot) processYouTube(msg *tgbotapi.Update) {
	u, err := url.Parse(msg.Message.Text)
	if err != nil {
		o.sendTelegramMessageToLog("error parse url", err.Error())
		return
	}
	if u.Host != "youtu.be" {
		o.sendTelegramMessageToLog("unknown host", u.Host)
		return
	}
	o.botAPI.DeleteMessage(msg.Message.Chat.ID, msg.Message.MessageID)
	playID := strings.ReplaceAll(u.Path, "/", "")

	o.botAPI.SendMessagePrepared(&utils.TelegramSendMessage{
		ChatID: msg.Message.Chat.ID,
		Text:   msg.Message.Text,
		InlineButtons: &[][]utils.TelegramInlineKeyboard{
			{
				o.createButton("<<", BACK_MORE, playID),
				o.createButton("<", BACK, playID),
				o.createButton(">", FORWARD, playID),
				o.createButton(">>", FORWARD_MORE, playID),
			},
			{
				o.createButton("<<", BACK_MORE, playID),
				o.createButton("<", BACK, playID),
				o.createButton(">", FORWARD, playID),
				o.createButton(">>", FORWARD_MORE, playID),
			},
			{
				o.createButton("0.75", SPEED_0_75, playID),
				o.createButton("1", SPEED_1, playID),
				o.createButton("1.25", SPEED_1_25, playID),
				o.createButton("1.5", SPEED_1_5, playID),
				o.createButton("1.75", SPEED_1_75, playID),
				o.createButton("2", SPEED_2, playID),
			},
			{
				o.createButton("play", PLAY, playID),
				o.createButton("close", CLOSE, playID),
			},
		},
	})
}

func (o *ControllerBot) createButton(text, code, playID string) utils.TelegramInlineKeyboard {
	return utils.TelegramInlineKeyboard{Text: text, CallbackData: "/" + code + "_" + playID}
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
