package bot_list

import (
	"container/list"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/abergasov/go_telegram_bots/pkg/config"
	"github.com/abergasov/go_telegram_bots/pkg/logger"
	"github.com/abergasov/go_telegram_bots/pkg/utils"
	"go.uber.org/zap"
	tgbotapi "gopkg.in/go-telegram-bot-api/telegram-bot-api.v4"
)

const (
	SPEED_LESS        = "a"
	SPEED_UP          = "b"
	FORWARD           = "g"
	FORWARD_MORE      = "h"
	BACK              = "i"
	BACK_MORE         = "j"
	PLAY              = "k"
	CLOSE             = "l"
	FULL_SCREEN       = "f"
	VOLUME_UP         = "v"
	VOLUME_DOWN       = "x"
	VOLUME_MUTE       = "m"
	REBOOT            = "r"
	commandBufferSize = 10
)

type CommandContainer struct {
	streamID string
	ch       chan Command
}

type Command struct {
	Cmd      string // команда для выполнения
	ActionID string // для какого видео
}

type ControllerBot struct {
	abstractBot
	muCommand *sync.RWMutex
	chatMap   map[int64]*list.List //chan Command
}

func NewControllerBot(conf *config.AppConfig, tWorker ITelegramWorker, buildTime, buildHash string) *ControllerBot {
	botName := "@rmcpi_bot"
	mB := &ControllerBot{
		muCommand: &sync.RWMutex{},
		chatMap:   make(map[int64]*list.List, 4),
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

func (o *ControllerBot) GetControlChan(targetChat int64, streamID string) chan Command {
	ch := make(chan Command, commandBufferSize)
	o.checkChanExit(targetChat)
	o.muCommand.Lock()
	o.chatMap[targetChat].PushBack(CommandContainer{ch: ch, streamID: streamID})
	o.muCommand.Unlock()
	return ch
}

func (o *ControllerBot) RemoveControlChan(chatID int64, streamID string) {
	o.checkChanExit(chatID)
	o.muCommand.Lock()
	logger.Info(fmt.Sprintf("try remove command listener. stream %s, chat %d", streamID, chatID))
	var el *list.Element
	for e := o.chatMap[chatID].Front(); e != nil; e = e.Next() {
		if e.Value.(CommandContainer).streamID == streamID {
			el = e
			logger.Info(fmt.Sprintf("found node listener. stream %s, chat %d", streamID, chatID))
			break
		}
	}
	o.chatMap[chatID].Remove(el)
	o.muCommand.Unlock()
	logger.Info(fmt.Sprintf("removed node listener. stream %s, chat %d", streamID, chatID))
}

func (o *ControllerBot) sendCommand(chatID int64, cmd Command) {
	o.checkChanExit(chatID)
	o.muCommand.RLock()
	for e := o.chatMap[chatID].Front(); e != nil; e = e.Next() {
		e.Value.(CommandContainer).ch <- cmd
	}
	o.muCommand.RUnlock()
}

func (o *ControllerBot) HandleRequest(msg *tgbotapi.Update) {
	go o.LogEvent(msg)
	if msg.Message != nil {
		switch msg.Message.Text {
		case "/start", "/help":
			o.sendBotInfo(msg)
			return
		case "reboot", "Reboot":
			o.sendCommand(msg.Message.Chat.ID, Command{Cmd: REBOOT, ActionID: ""})
			return
		}
		o.processYouTube(msg)
	}
	if msg.CallbackQuery != nil {
		o.handleCallback(msg.CallbackQuery.Data, msg.CallbackQuery.ID, msg.CallbackQuery.Message.Chat.ID)
	}
}

func (o *ControllerBot) handleCallback(callbackMessage, callbackQueryID string, chatID int64) {
	msg := strings.ReplaceAll(callbackMessage, "/", "")
	data := strings.Split(msg, "_")
	o.sendCommand(chatID, Command{Cmd: data[0], ActionID: data[1]})
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
	o.checkChanExit(msg.Message.Chat.ID)
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
	o.botAPI.SendMessage(msg.Message.Chat.ID, "Bot allow control pi. chat: "+strconv.FormatInt(msg.Message.Chat.ID, 10), nil)
}

func (o *ControllerBot) checkChanExit(chatID int64) {
	o.muCommand.RLock()
	_, ok := o.chatMap[chatID]
	o.muCommand.RUnlock()
	if !ok {
		o.muCommand.Lock()
		o.chatMap[chatID] = list.New()
		o.muCommand.Unlock()
	}
	for e := o.chatMap[chatID].Front(); e != nil; e = e.Next() {
		o.eraseChan(e.Value.(CommandContainer))
	}
}

func (o *ControllerBot) eraseChan(c CommandContainer) {
	if len(c.ch) < cap(c.ch) {
		return
	}
	logger.Info("channel is full, erase it")
	var cmd Command
	for len(c.ch) > 0 {
		cmd = <-c.ch
		logger.Info("clear unread command", zap.String("id", cmd.ActionID), zap.String("cmd", cmd.Cmd))
	}
	logger.Info("channel erased")
}
