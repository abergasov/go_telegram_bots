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
	"strconv"
	"strings"
	"sync"
)

const (
	SPEED_LESS   = "a"
	SPEED_UP     = "b"
	FORWARD      = "g"
	FORWARD_MORE = "h"
	BACK         = "i"
	BACK_MORE    = "j"
	PLAY         = "k"
	CLOSE        = "l"
	FULL_SCREEN  = "f"
	VOLUME_UP    = "v"
	VOLUME_DOWN  = "x"
	VOLUME_MUTE  = "m"
	REBOOT       = "r"
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
		case "reboot", "Reboot":
			o.MuCommand.Lock()
			o.ActiveCommand = &Command{
				Cmd:      REBOOT,
				ActionID: "",
			}
			o.MuCommand.Unlock()
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
	validHost := u.Host == "youtu.be" || u.Host == "youtube.com" || u.Host == "www.youtube.com"
	if !validHost {
		o.sendTelegramMessageToLog("unknown host", u.Host)
		return
	}
	o.botAPI.DeleteMessage(msg.Message.Chat.ID, msg.Message.MessageID)
	playID := strings.ReplaceAll(u.Path, "/", "")
	if playID == "watch" {
		tmp := strings.Split(strings.ReplaceAll(u.RawQuery, "v=", ""), "&")
		playID = tmp[0]
	}
	playID = strings.ReplaceAll(playID, "_", "%")

	dataPosition := make([]utils.TelegramInlineKeyboard, 0, 8)
	for i := 0; i < 8; i++ {
		cmd := strconv.Itoa(i)
		dataPosition = append(dataPosition, o.createButton(cmd, cmd, playID))
	}

	o.botAPI.SendMessagePrepared(&utils.TelegramSendMessage{
		ChatID: msg.Message.Chat.ID,
		Text:   msg.Message.Text,
		InlineButtons: &[][]utils.TelegramInlineKeyboard{
			dataPosition,
			{
				o.createButton("<<", BACK_MORE, playID),
				o.createButton("<", BACK, playID),
				o.createButton(">", FORWARD, playID),
				o.createButton(">>", FORWARD_MORE, playID),
			},
			{
				o.createButton("<-", SPEED_LESS, playID),
				o.createButton("->", SPEED_UP, playID),
			},
			{
				o.createButton("+", VOLUME_UP, playID),
				o.createButton("-", VOLUME_DOWN, playID),
				o.createButton("m", VOLUME_MUTE, playID),
			},
			{
				o.createButton("play", PLAY, playID),
				o.createButton("f", FULL_SCREEN, playID),
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
