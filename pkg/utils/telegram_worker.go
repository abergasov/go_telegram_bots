package utils

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/abergasov/go_telegram_bots/pkg/logger"

	"go.uber.org/zap"
)

type TelegramWorker struct {
	token   string
	baseURL string
}

func NewWorker() *TelegramWorker {
	return &TelegramWorker{}
}

func (t *TelegramWorker) SetToken(token string) {
	t.token = token
	t.baseURL = "https://api.telegram.org/bot" + t.token + "/"
}

func (t *TelegramWorker) DeleteMessage(chatID int64, messageID int) {
	requestBody, _ := json.Marshal(map[string]interface{}{
		"chat_id":    chatID,
		"message_id": messageID,
	})
	t.sendTelegramRequest(requestBody, "deleteMessage")
}

func (t *TelegramWorker) SendAlert(alm *TelegramSendAlert) {
	requestBody, _ := json.Marshal(&alm)
	t.sendTelegramRequest(requestBody, "answerCallbackQuery")
}

func (t *TelegramWorker) SendMessage(chat int64, message string, replyMarkup *TelegramKeyboard, params ...string) {
	tm := TelegramSendMessage{
		ChatID:    chat,
		Text:      t.prepareMessage(message, params...),
		ParseMode: "html",
	}
	if replyMarkup != nil {
		tm.ReplyMarkup = replyMarkup
	}
	t.SendMessagePrepared(&tm)
}

func (t *TelegramWorker) SendMessagePrepared(tm *TelegramSendMessage) {
	path := t.getAPIPath(tm)
	if tm.InlineButtons != nil && len(*tm.InlineButtons) > 0 {
		tm.ReplyMarkup = map[string]*[][]TelegramInlineKeyboard{
			"inline_keyboard": tm.InlineButtons,
		}
	} else if tm.audioPack != nil && len(*tm.audioPack) > 0 {
		tm.Media = tm.audioPack
	}
	requestBody, _ := json.Marshal(tm)
	t.sendTelegramRequest(requestBody, path)
}

func (t *TelegramWorker) sendTelegramRequest(requestBody []byte, path string) {
	resp, err := http.Post(t.baseURL+path, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		logger.Error("error send telegram message", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, errR := ioutil.ReadAll(resp.Body)
		if errR != nil {
			logger.Error("error read telegram response", errR)
		} else {
			logger.Warning("error from telegram", zap.String("response", string(body)), zap.String("request", string(requestBody)))
		}
	}
}

func (t *TelegramWorker) prepareMessage(message string, params ...string) string {
	values := make([]string, 0, len(params)+1)
	values = append(values, message)
	for i := range params {
		values = append(values, params[i])
	}
	return strings.Join(values, "\n")
}

func (t *TelegramWorker) getAPIPath(tm *TelegramSendMessage) string {
	path := "sendMessage"
	if len(tm.Photo) > 0 {
		return "sendPhoto"
	} else if tm.audioPack != nil && len(*tm.audioPack) > 0 {
		return "sendMediaGroup"
	} else if tm.Audio != "" {
		return "sendAudio"
	}
	return path
}
