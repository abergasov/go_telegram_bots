package bot_manager

import (
	"github.com/abergasov/go_telegram_bots/pkg/bot_manager/bot_list"
	"github.com/abergasov/go_telegram_bots/pkg/config"
	"github.com/abergasov/go_telegram_bots/pkg/logger"
	"github.com/abergasov/go_telegram_bots/pkg/utils"

	tgbotapi "gopkg.in/go-telegram-bot-api/telegram-bot-api.v4"

	"go.uber.org/zap"
)

type SingleBot interface {
	GetMe() string
	GetHookPath(string, string) (string, string)
	HandleRequest(*tgbotapi.Update)
	IsAdminChat(*tgbotapi.Update) bool
	LogEvent(*tgbotapi.Update)
}

type BotManager struct {
	botPool      []SingleBot
	HooksURLList map[string]string
}

func InitBotManager(config *config.AppConfig, hookPostfix, buildTime, buildHash string) *BotManager {
	bM := &BotManager{
		botPool: []SingleBot{
			bot_list.NewMusicBot(config, utils.NewWorker(), buildTime, buildHash),
			bot_list.NewOrchestraBot(config, utils.NewWorker(), buildTime, buildHash),
			bot_list.NewControllerBot(config, utils.NewWorker(), buildTime, buildHash),
			bot_list.NewBelformagBot(config, utils.NewWorker(), buildTime, buildHash),
		},
	}
	bM.getHookUrls(config.HostURL, hookPostfix)
	return bM
}

func (b *BotManager) GetBot(botName string) SingleBot {
	for i := range b.botPool {
		if b.botPool[i].GetMe() != botName {
			continue
		}
		return b.botPool[i]
	}
	return nil
}

func (b *BotManager) HandleTelegramRequest(msg *tgbotapi.Update, botName string) {
	for i := range b.botPool {
		if b.botPool[i].GetMe() != botName {
			continue
		}
		b.botPool[i].HandleRequest(msg)
		if !b.botPool[i].IsAdminChat(msg) {
			b.botPool[i].LogEvent(msg)
		}
		//b.botPool[i].LogEvent(msg)
	}
}

func (b *BotManager) getHookUrls(hostURL, hookPostfix string) {
	b.HooksURLList = make(map[string]string)
	for i := range b.botPool {
		botName, botHook := b.botPool[i].GetHookPath(hostURL, hookPostfix)
		logger.Info("hook url", zap.String(botName, botHook))
		b.HooksURLList[botHook] = botName
	}
}
