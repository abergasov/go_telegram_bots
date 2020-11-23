package bot_list

import (
	"bytes"
	"encoding/json"
	"errors"
	"go_bots/pkg/config"
	"go_bots/pkg/logger"
	"net/http"
	"strconv"
	"strings"

	tgbotapi "gopkg.in/go-telegram-bot-api/telegram-bot-api.v4"
)

type channelData struct {
	Title       string       `json:"title"`
	ID          int64        `json:"id"`
	UserName    string       `json:"user_name"`
	Description string       `json:"description"`
	InviteLink  string       `json:"invite_link"`
	ProxyList   []*proxyData `json:"proxy_list"`
}

type OrchestraBot struct {
	orchestraURL string
	orchestraKey string
	abstractBot
}

type proxyData struct {
	Server string `json:"server"`
	Port   string `json:"port"`
	Secret string `json:"secret"`
}

func NewOrchestraBot(conf *config.AppConfig, tWorker ITelegramWorker, buildTime, buildHash string) *OrchestraBot {
	botName := "@short_news_helper_bot"
	oB := &OrchestraBot{
		orchestraURL: conf.OrchestraUrl,
		orchestraKey: conf.OrchestraKey,
	}
	for i := range conf.BotList {
		if conf.BotList[i].BotName != botName {
			continue
		}
		oB.InitBot(tWorker, botName, &conf.BotList[i])
		break
	}
	if !oB.initialized {
		logger.Error("short_news_helper_bot is not initializes", errors.New("config missing"))
	} else {
		oB.sendTelegramMessageToAdmin("short_news_helper_bot started", buildHash, buildTime)
	}
	return oB
}

func (o *OrchestraBot) HandleRequest(msg *tgbotapi.Update) {
	isForwardFromChannel := msg.Message != nil &&
		msg.Message.ForwardFromChat != nil &&
		msg.Message.ForwardFromChat.Type == "channel"
	if !isForwardFromChannel {
		o.botAPI.SendMessage(msg.Message.Chat.ID, "это не канал", nil)
		return
	}
	o.collectChanel(msg)
}

func (o *OrchestraBot) proxySearch(msg *tgbotapi.Update) []*proxyData {
	pdList := make([]*proxyData, 0, 10)
	if strings.Contains(msg.Message.Text, `https://t.me/proxy?`) {
		// mtp proxy here
		p := o.collectProxy(msg.Message.Text)
		if p != nil {
			pdList = append(pdList, p)
		}
	}
	if msg.Message.Entities != nil && len(*msg.Message.Entities) > 0 {
		for _, i := range *msg.Message.Entities {
			if !strings.Contains(i.URL, `https://t.me/proxy?`) {
				continue
			}
			// mtp proxy here
			p := o.collectProxy(i.URL)
			if p != nil {
				pdList = append(pdList, p)
			}
		}
	}
	if len(pdList) > 0 {
		o.botAPI.SendMessage(msg.Message.Chat.ID, "найдены прокси", nil, strconv.Itoa(len(pdList)))
	}
	return pdList
}

func (o *OrchestraBot) collectProxy(text string) *proxyData {
	list := strings.Split(text, `https://t.me/proxy?`)
	if len(list) != 2 {
		return nil
	}
	pData := strings.Split(list[1], `&`)
	if len(pData) != 3 {
		return nil
	}

	pD := &proxyData{}
	for i := range pData {
		cD := strings.Split(pData[i], `=`)
		if len(cD) != 2 {
			continue
		}
		switch cD[0] {
		case "server":
			pD.Server = cD[1]
		case "port":
			pD.Port = cD[1]
		case "secret":
			pD.Secret = cD[1]
		}
	}
	return pD
}

func (o *OrchestraBot) collectChanel(msg *tgbotapi.Update) {
	proxyList := make([]*proxyData, 0, 10)
	if msg.Message != nil {
		proxyList = o.proxySearch(msg)
	}
	data, err := json.Marshal(channelData{
		ID:          msg.Message.ForwardFromChat.ID,
		UserName:    msg.Message.ForwardFromChat.UserName,
		Title:       msg.Message.ForwardFromChat.Title,
		Description: msg.Message.ForwardFromChat.Description,
		InviteLink:  msg.Message.ForwardFromChat.InviteLink,
		ProxyList:   proxyList,
	})
	if err != nil {
		o.botAPI.SendMessage(msg.Message.Chat.ID, "ошибка отправки данных", nil, err.Error())
		logger.Error("error decode data 4 external", err)
		return
	}

	if err = o.sendData(data); err != nil {
		logger.Error("error send data to external", err)
		o.botAPI.SendMessage(msg.Message.Chat.ID, "ошибка отправки данных", nil, err.Error())
		return
	}
	o.botAPI.SendMessage(msg.Message.Chat.ID, "Данные сохранены", nil, msg.Message.ForwardFromChat.Title)
}

func (o *OrchestraBot) sendData(data []byte) error {
	client := &http.Client{}
	req, _ := http.NewRequest(http.MethodPost, o.orchestraURL, bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Token", o.orchestraKey)
	req.Header.Set("Connection", "keep-alive")
	response, err := client.Do(req)
	if err != nil {
		return err
	}
	if response != nil && response.StatusCode != 200 {
		defer response.Body.Close()
		return errors.New("invalid request: " + response.Status)
	}
	return nil
}

func (o *OrchestraBot) LogEvent(msg *tgbotapi.Update) {
	data, err := json.Marshal(msg)
	if err != nil {
		logger.Error("error log event", err)
		return
	}
	o.sendTelegramMessageToLog(string(data))
}
