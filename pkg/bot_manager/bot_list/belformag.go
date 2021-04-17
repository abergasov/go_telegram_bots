package bot_list

import (
	"bytes"
	"encoding/json"
	"errors"
	"go_telegram_bots/pkg/config"
	"go_telegram_bots/pkg/logger"
	"go_telegram_bots/pkg/utils"
	"io/ioutil"
	"net/http"
	"strings"

	tgbotapi "gopkg.in/go-telegram-bot-api/telegram-bot-api.v4"
)

type BelformagBot struct {
	urlPath  string
	urlToken string
	abstractBot
}

type appResp struct {
	OK   bool              `json:"ok"`
	Shop string            `json:"shop"`
	Apps map[string]string `json:"apps"`
}

func NewBelformagBot(conf *config.AppConfig, tWorker ITelegramWorker, buildTime, buildHash string) *BelformagBot {
	botName := "@belformag_bot"
	bf := &BelformagBot{
		urlPath:  conf.BmxURL,
		urlToken: conf.BmxKey,
	}
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
		commands := strings.Split(msg.CallbackQuery.Data, "_")
		if len(commands) != 3 {
			return
		}
		switch commands[0] {
		case "/d":
			b.cancelSub(commands[2], commands[1], msg)
		}
	}
}

func (b *BelformagBot) cancelSub(shop, app string, msg *tgbotapi.Update) {
	b.botAPI.SendMessage(msg.CallbackQuery.Message.Chat.ID, "Start cancelling", nil, "")
	resp, err := b.askBmx(http.MethodPost, shop, app)
	if err != nil {
		b.botAPI.SendMessage(msg.CallbackQuery.Message.Chat.ID, "Error cancel sub for shop", nil, err.Error())
		return
	}
	if resp.OK {
		b.botAPI.DeleteMessage(msg.CallbackQuery.Message.Chat.ID, msg.CallbackQuery.Message.MessageID)
		return
	}
}

func (b *BelformagBot) processCommand(msg *tgbotapi.Update) {
	if !strings.Contains(msg.Message.Text, ".myshopify.com") {
		return
	}
	resp, err := b.askBmx(http.MethodGet, msg.Message.Text, "")
	if err != nil {
		b.botAPI.SendMessage(msg.Message.Chat.ID, "Error ask for shop", nil, err.Error())
		return
	}

	inlineKeyboards := make([][]utils.TelegramInlineKeyboard, 0, len(resp.Apps))
	for i := range resp.Apps {
		inlineKeyboards = append(inlineKeyboards, []utils.TelegramInlineKeyboard{
			{
				Text:         resp.Apps[i],
				CallbackData: "/d_" + i + "_" + resp.Shop,
			},
		})
	}
	b.botAPI.SendMessagePrepared(&utils.TelegramSendMessage{
		ChatID:        msg.Message.Chat.ID,
		Text:          "found apps for shop " + msg.Message.Text,
		InlineButtons: &inlineKeyboards,
	})
}

func (b *BelformagBot) askBmx(method, shopDomain, app string) (*appResp, error) {
	var str []byte
	client := &http.Client{}
	if method != http.MethodGet {

		str, _ = json.Marshal(map[string]string{"shop": shopDomain, "app": app})
	}
	req, _ := http.NewRequest(method, b.urlPath, bytes.NewBuffer(str))
	if method == http.MethodGet {
		q := req.URL.Query()
		q.Add("shop", shopDomain)
		q.Add("app", app)
		req.URL.RawQuery = q.Encode()
	} else {
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
	}

	req.Header.Set("X-Token", b.urlToken)
	response, err := client.Do(req)
	if err != nil {
		logger.Error("error ask", err)
		return nil, err
	}
	if response == nil {
		return nil, errors.New("empty response")
	}
	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	jsonString := string(body)
	okCondition := response.StatusCode == 200 || response.StatusCode == 201
	if !okCondition {
		err = errors.New(jsonString)
		logger.Error("error ask", err)
		return nil, err
	}
	resp := &appResp{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
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
