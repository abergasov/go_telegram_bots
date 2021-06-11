package controller

import "github.com/abergasov/go_telegram_bots/pkg/bot_manager/bot_list"

type CommandBot interface {
	GetControlChan(chatID int64, streamID string) chan bot_list.Command
	RemoveControlChan(chatID int64, streamID string)
}
