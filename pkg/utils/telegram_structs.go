package utils

type TelegramSendMessage struct {
	ChatID      int64       `json:"chat_id"`
	Text        string      `json:"text,omitempty"`
	Photo       string      `json:"photo,omitempty"`
	Caption     string      `json:"caption,omitempty"`
	ParseMode   string      `json:"parse_mode"`
	ReplyMarkup interface{} `json:"reply_markup,omitempty"`
	Media       interface{} `json:"media,omitempty"`
	Audio       string      `json:"audio,omitempty"`
	// keyboardButtons *TelegramKeyboard
	InlineButtons *[][]TelegramInlineKeyboard `json:"inline_buttons,omitempty"`
	audioPack     *[]TelegramAudio
}

type TelegramSendAlert struct {
	CallbackQueryID string `json:"callback_query_id"`
	Text            string `json:"text"`
	ShowAlert       bool   `json:"show_alert"`
}

type TelegramKeyboard struct {
	OneTimeKeyboard bool       `json:"one_time_keyboard"`
	Resize          bool       `json:"resize"`
	Keyboard        [][]string `json:"keyboard"`
}

type TelegramInlineKeyboard struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data"`
}

type TelegramAudio struct {
	Type      string `json:"type"`
	Media     string `json:"media"`
	Caption   string `json:"caption,omitempty"`
	Title     string `json:"title,omitempty"`
	Performer string `json:"performer,omitempty"`
}
