package bot_list

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/abergasov/go_telegram_bots/pkg/config"
	"github.com/abergasov/go_telegram_bots/pkg/logger"
	"github.com/abergasov/go_telegram_bots/pkg/utils"

	tgbotapi "gopkg.in/go-telegram-bot-api/telegram-bot-api.v4"
)

type MusicBot struct {
	abstractBot
}

type songsSend struct {
	fileID string
	title  string
	songID int
	album  string
	band   string
	year   string
}

type albumData struct {
	albumID int
	band    string
	album   string
	year    string
	photoID string
}

func NewMusicBot(conf *config.AppConfig, tWorker ITelegramWorker, buildTime, buildHash string) *MusicBot {
	botName := "@My_muisic_bot"
	mB := &MusicBot{}
	for i := range conf.BotList {
		if conf.BotList[i].BotName != botName {
			continue
		}
		mB.InitBot(tWorker, botName, &conf.BotList[i])
		break
	}
	if !mB.initialized {
		logger.Error("MusicBot is not initializes", errors.New("config missing"))
	} else {
		mB.sendTelegramMessageToAdmin("MusicBot started", buildHash, buildTime)
	}
	return mB
}

func (m *MusicBot) HandleRequest(msg *tgbotapi.Update) {
	if msg.Message != nil {
		switch msg.Message.Text {
		case "/start", "/help":
			m.sendBotInfo(msg)
			return
		case "/bands", "Bands", "bands":
			m.sendBandsList(msg)
			return
		}
	}

	if msg.CallbackQuery != nil {
		commandList := strings.Split(msg.CallbackQuery.Data, "_")
		if len(commandList) == 2 {
			switch commandList[0] {
			case "/da": // download album
				m.downloadAlbum(commandList[1], msg)
			case "/slk", "/sdlk": // like/dislike song
				m.rateSong(commandList, msg)
			case "/als": // all liked songs band
				m.bandLikedSongs(commandList[1], msg)
			case "/dal", "/dad": // download album liken|disliked songs
				m.albumRatedSongs(commandList[0], commandList[1], msg)
			case "/sa": // show album
				m.rotateAlbum(commandList[1], msg)
			}
			return
		}
	}
	m.search(msg)
}

func (m *MusicBot) LogEvent(msg *tgbotapi.Update) {
	data, err := json.Marshal(msg)
	if err != nil {
		logger.Error("error log event", err)
		return
	}
	m.sendTelegramMessageToLog(string(data))
}

func (m *MusicBot) sendBotInfo(msg *tgbotapi.Update) {
	m.botAPI.SendMessage(msg.Message.Chat.ID, "Bot keep and show music", nil, "List of groups /bands", "Show menu /menu")
}

func (m *MusicBot) sendBandsList(msg *tgbotapi.Update) {
	rows, err := m.db.SelectQuery("SELECT band_name FROM bands ORDER BY band_name")
	if err != nil {
		logger.Error("error load bands", err)
	}
	bands := make([][]string, 0, 20)
	bandsRow := make([]string, 0, 3)
	counter := 0
	for rows.Next() {
		if counter > 2 {
			bands = append(bands, bandsRow)
			counter = 0
			bandsRow = make([]string, 0, 3)
		}
		var band string
		err = rows.Scan(&band)
		if err != nil {
			continue
		}
		bandsRow = append(bandsRow, band)
		counter++
	}
	if counter > 0 {
		bands = append(bands, bandsRow)
	}
	m.botAPI.SendMessage(msg.Message.Chat.ID, "Select band", &utils.TelegramKeyboard{
		OneTimeKeyboard: true,
		Resize:          false,
		Keyboard:        bands,
	})
}

func (m *MusicBot) search(msg *tgbotapi.Update) {
	sqlB := "SELECT a_id, b_id, band_name, album_name, label, year FROM bands m LEFT JOIN albums a ON a.band_id = m.b_id WHERE m.band_name = ? ORDER BY a.year"
	rows, err := m.db.SelectQuery(sqlB, msg.Message.Text)
	if err != nil {
		logger.Error("error load bands from db", err)
		return
	}
	defer rows.Close()

	albumList := make([]*albumData, 0, 15)
	var bandID string
	counter := 0
	for rows.Next() {
		counter++
		var a albumData
		err = rows.Scan(&a.albumID, &bandID, &a.band, &a.album, &a.photoID, &a.year)
		if err != nil {
			continue
		}
		albumList = append(albumList, &a)
	}
	if counter == 0 {
		m.searchSong(msg)
		return
	}

	inlineKeyboards := make([][]utils.TelegramInlineKeyboard, 0, 3)

	albumIDStr := strconv.Itoa(albumList[0].albumID)
	albumPagination := make([]utils.TelegramInlineKeyboard, 0, counter)
	title := fmt.Sprintf("1 of %d \n %s - %s (%s)", counter, albumList[0].band, albumList[0].album, albumList[0].year)
	photoID := albumList[0].photoID
	for i := 1; i <= counter; i++ {
		text := strconv.Itoa(i)
		if i == 1 {
			text = ">" + text + "<"
		}
		if i == 8 {
			text = "8 >"
		}
		albumPagination = append(albumPagination, utils.TelegramInlineKeyboard{
			Text:         text,
			CallbackData: "/sa_" + strconv.Itoa(albumList[i-1].albumID),
		})
		if i == 8 {
			break
		}
	}
	inlineKeyboards = append(inlineKeyboards, albumPagination, []utils.TelegramInlineKeyboard{
		{
			Text:         "Full album",
			CallbackData: "/da_" + albumIDStr,
		},
		{
			Text:         "\xF0\x9F\x91\x8D",
			CallbackData: "/dal_" + albumIDStr,
		},
		{
			Text:         "\xF0\x9F\x91\x8E",
			CallbackData: "/dad_" + albumIDStr,
		},
	}, []utils.TelegramInlineKeyboard{
		{
			Text:         "All band liked songs",
			CallbackData: "/als_" + bandID,
		},
	})

	m.botAPI.SendMessagePrepared(&utils.TelegramSendMessage{
		ChatID:        msg.Message.Chat.ID,
		Photo:         photoID,
		Caption:       title,
		InlineButtons: &inlineKeyboards,
	})
}

func (m *MusicBot) searchSong(msg *tgbotapi.Update) {
	sqlS := "SELECT m.song, b.band_name, a.album_name, a.year, m.id FROM muisic m LEFT JOIN bands b ON m.band = b.b_id LEFT JOIN albums a ON a.a_id = m.album WHERE m.song LIKE ? ORDER BY a.year"
	rows, err := m.db.SelectQuery(sqlS, "%"+msg.Message.Text+"%")
	if err != nil {
		logger.Error("error load songs from db", err)
		return
	}
	defer rows.Close()
	songsList := make([]string, 0, 20)
	for rows.Next() {
		var song, band, album, year string
		var songID int
		err = rows.Scan(&song, &band, &album, &year, &songID)
		if err != nil {
			continue
		}
		songsList = append(songsList, fmt.Sprintf("/%d %s - %s from %s (%s)", songID, song, band, album, year))
	}
	if len(songsList) == 0 {
		m.botAPI.SendMessage(msg.Message.Chat.ID, "there is no songs here", nil)
		return
	}
	m.botAPI.SendMessage(msg.Message.Chat.ID, strings.Join(songsList, "\n"), nil)
}

func (m *MusicBot) bandLikedSongs(bandID string, msg *tgbotapi.Update) {
	sqlS := "SELECT m.file_id, m.song, b.band_name, a.album_name, a.year, m.id FROM muisic m LEFT JOIN bands b ON m.band = b.b_id LEFT JOIN albums a ON a.a_id = m.album WHERE m.band = ? ORDER BY a.year"
	rows, err := m.db.SelectQuery(sqlS, bandID)
	if err != nil {
		logger.Error("error load songs from db", err)
		return
	}
	defer rows.Close()
	songs := make([]*songsSend, 0, 20)
	for rows.Next() {
		song := &songsSend{}
		err := rows.Scan(&song.fileID, &song.title, &song.band, &song.album, &song.year, &song.songID)
		if err != nil {
			continue
		}
		songs = append(songs, song)
	}
	m.sendSongs(songs, int64(msg.CallbackQuery.From.ID))
}

func (m *MusicBot) albumRatedSongs(rateType, albumID string, msg *tgbotapi.Update) {
	var sqlS string
	if rateType == "/dal" {
		sqlS = "SELECT m.song, m.file_id, m.id FROM muisic m LEFT JOIN like_songs ls on m.id = ls.song_id AND ls.user_id = ? WHERE album = ? AND ls.user_id IS NOT NULL"
	} else {
		sqlS = "SELECT m.song, m.file_id, m.id FROM muisic m LEFT JOIN dislike_songs ls on m.id = ls.song_id AND ls.user_id = ? WHERE album = ? AND ls.user_id IS NOT NULL"
	}
	rows, err := m.db.SelectQuery(sqlS, msg.CallbackQuery.From.ID, albumID)
	if err != nil {
		logger.Error("error load songs from db", err)
		return
	}
	defer rows.Close()
	songs := make([]*songsSend, 0, 20)
	for rows.Next() {
		song := &songsSend{}
		err := rows.Scan(&song.title, &song.fileID, &song.songID)
		if err != nil {
			continue
		}
		songs = append(songs, song)
	}
	m.sendSongs(songs, int64(msg.CallbackQuery.From.ID))
}

func (m *MusicBot) sendSongs(songs []*songsSend, chatID int64) {
	for i := range songs {
		caption := songs[i].title
		if songs[i].band != "" {
			caption = fmt.Sprintf("%s - %s from %s (%s)", songs[i].title, songs[i].band, songs[i].album, songs[i].year)
		}
		songIDString := strconv.Itoa(songs[i].songID)
		m.botAPI.SendMessagePrepared(&utils.TelegramSendMessage{
			ChatID:  chatID,
			Audio:   songs[i].fileID,
			Caption: caption,
			InlineButtons: &[][]utils.TelegramInlineKeyboard{
				{
					utils.TelegramInlineKeyboard{Text: "\xF0\x9F\x91\x8D", CallbackData: "/slk_" + songIDString},
					utils.TelegramInlineKeyboard{Text: "\xF0\x9F\x91\x8E", CallbackData: "/sdlk_" + songIDString},
				},
			},
		})
	}
}

func (m *MusicBot) downloadAlbum(albumID string, msg *tgbotapi.Update) {
	rows, err := m.db.SelectQuery("SELECT file_id, song, id FROM muisic WHERE album = ? ORDER BY id", albumID)
	if err != nil {
		logger.Error("error select album", err)
		return
	}
	defer rows.Close()
	songs := make([]*songsSend, 0, 20)
	for rows.Next() {
		song := &songsSend{}
		err := rows.Scan(&song.fileID, &song.title, &song.songID)
		if err != nil {
			continue
		}
		songs = append(songs, song)
	}
	m.sendSongs(songs, msg.CallbackQuery.Message.Chat.ID)
}

func (m *MusicBot) rateSong(commandList []string, msg *tgbotapi.Update) {
	var sqlR, sqlD, alertMsg string
	if commandList[0] == "/slk" {
		sqlR = "REPLACE INTO like_songs SET user_id = ?, song_id = ?"
		sqlD = "DELETE FROM dislike_songs WHERE user_id = ? AND song_id = ?"
		alertMsg = "You like this song"
	} else {
		sqlR = "REPLACE INTO dislike_songs SET user_id = ?, song_id = ?"
		sqlD = "DELETE FROM like_songs WHERE user_id = ? AND song_id = ?"
		alertMsg = "You dislike this song"
		m.botAPI.DeleteMessage(msg.CallbackQuery.Message.Chat.ID, msg.CallbackQuery.Message.MessageID)
	}
	m.db.Exec(sqlR, msg.CallbackQuery.Message.Chat.ID, commandList[1])
	m.db.Exec(sqlD, msg.CallbackQuery.Message.Chat.ID, commandList[1])
	m.botAPI.SendAlert(&utils.TelegramSendAlert{
		Text:            alertMsg,
		ShowAlert:       false,
		CallbackQueryID: msg.CallbackQuery.ID,
	})
}

func (m *MusicBot) rotateAlbum(albumID string, msg *tgbotapi.Update) {
	sqlS := "SELECT a_id, band_id, band_name, album_name, label, year FROM albums a LEFT JOIN bands b ON b.b_id = a.band_id WHERE b.b_id = (SELECT az.band_id FROM albums az WHERE az.a_id = ?) ORDER BY a.year"
	rows, err := m.db.SelectQuery(sqlS, albumID)
	if err != nil {
		logger.Error("error load bands from db 4 rotate", err)
		return
	}
	defer rows.Close()

	albumPagination := make([]utils.TelegramInlineKeyboard, 0, 30)
	counter := 0
	foundAtPosition := 0
	var photoID, title string
	var bandID int
	for rows.Next() {
		var band, year, albumName, photo string
		var aID int
		err = rows.Scan(&aID, &bandID, &band, &albumName, &photo, &year)
		if err != nil {
			continue
		}
		counter++
		aIDStr := strconv.Itoa(aID)
		counterStr := strconv.Itoa(counter)
		text := counterStr
		if albumID == aIDStr {
			text = ">" + counterStr + "<"
		}
		albumPagination = append(albumPagination, utils.TelegramInlineKeyboard{
			Text:         text,
			CallbackData: "/sa_" + aIDStr,
		})

		if albumID != aIDStr {
			continue
		}
		foundAtPosition = counter
		photoID = photo
		title = fmt.Sprintf("%s - %s (%s)", band, albumName, year)
	}
	if len(albumPagination) > 8 {
		// в дело пошла высшая математика (с)
		beginSlice := foundAtPosition - 3
		if foundAtPosition-3 < 0 {
			beginSlice = 0
		}
		endSlice := beginSlice + 8
		if endSlice > counter {
			endSlice = counter
			if endSlice-8 < 0 {
				beginSlice = 0
			} else {
				beginSlice = endSlice - 8
			}
		}
		albumPagination = albumPagination[beginSlice:endSlice]
		if beginSlice > 0 {
			albumPagination[0].Text = "< " + albumPagination[0].Text
		}
		if endSlice < counter {
			albumPagination[7].Text += " >"
		}
	}
	inlineKeyboards := make([][]utils.TelegramInlineKeyboard, 0, 3)

	inlineKeyboards = append(inlineKeyboards, albumPagination, []utils.TelegramInlineKeyboard{
		{
			Text:         "Full album",
			CallbackData: "/da_" + albumID,
		},
		{
			Text:         "\xF0\x9F\x91\x8D",
			CallbackData: "/dal_" + albumID,
		},
		{
			Text:         "\xF0\x9F\x91\x8E",
			CallbackData: "/dad_" + albumID,
		},
	}, []utils.TelegramInlineKeyboard{
		{
			Text:         "All band liked songs",
			CallbackData: "/als_" + strconv.Itoa(bandID),
		},
	})
	m.botAPI.SendMessagePrepared(&utils.TelegramSendMessage{
		ChatID:        int64(msg.CallbackQuery.From.ID),
		Photo:         photoID,
		Caption:       title,
		InlineButtons: &inlineKeyboards,
	})
	m.botAPI.DeleteMessage(msg.CallbackQuery.Message.Chat.ID, msg.CallbackQuery.Message.MessageID)
}
