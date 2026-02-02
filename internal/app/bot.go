package app

import (
	"log"
	"sort"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (a *App) RunBot(ctxDone <-chan struct{}) {
	updates := a.bot.GetUpdatesChan(tgbotapi.NewUpdate(0))
	for {
		select {
		case <-ctxDone:
			return
		case update := <-updates:
			a.handleUpdate(update)
		}
	}
}

func (a *App) handleUpdate(update tgbotapi.Update) {
	if update.Message != nil {
		chatID := update.Message.Chat.ID
		a.rememberChat(chatID)
		statuses := a.collectStatuses()
		for _, status := range statuses {
			msg := tgbotapi.NewMessage(chatID, renderRoomStatus(status))
			msg.ReplyMarkup = roomKeyboard(status.Room.ID)
			if _, err := a.bot.Send(msg); err != nil {
				log.Printf("telegram send: %v", err)
			}
		}
		return
	}
	if update.CallbackQuery != nil {
		cb := update.CallbackQuery
		a.rememberChat(cb.Message.Chat.ID)
		a.handleCallback(cb)
	}
}

func (a *App) handleCallback(cb *tgbotapi.CallbackQuery) {
	parts := strings.Split(cb.Data, ":")
	if len(parts) != 3 || parts[0] != "room" {
		return
	}
	roomID := parts[1]
	action := parts[2]
	room, ok := a.findRoom(roomID)
	if !ok {
		return
	}

	turnOn := action == "on"
	if err := a.publishThermostat(room, turnOn); err != nil {
		log.Printf("publish thermostat: %v", err)
	}
	a.storeManualOverride(room.ID)
	answer := tgbotapi.NewCallback(cb.ID, "Команда отправлена")
	if _, err := a.bot.Request(answer); err != nil {
		log.Printf("callback answer: %v", err)
	}
}

func (a *App) rememberChat(chatID int64) {
	a.chatMu.Lock()
	a.chatIDs[chatID] = struct{}{}
	a.chatMu.Unlock()
}

func (a *App) broadcast(message string) {
	if a.cfg.ChatID == 0 {
		return
	}
	msg := tgbotapi.NewMessage(a.cfg.ChatID, message)
	if _, err := a.bot.Send(msg); err != nil {
		log.Printf("telegram broadcast: %v", err)
	}
}

func (a *App) listChats() []int64 {
	a.chatMu.RLock()
	defer a.chatMu.RUnlock()
	ids := make([]int64, 0, len(a.chatIDs))
	for id := range a.chatIDs {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

func roomKeyboard(roomID string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Включить", "room:"+roomID+":on"),
			tgbotapi.NewInlineKeyboardButtonData("Выключить", "room:"+roomID+":off"),
		),
	)
}

func renderRoomStatus(status RoomStatus) string {
	builder := &strings.Builder{}
	builder.WriteString(status.Room.Name)
	builder.WriteString(":\n")

	keys := make([]string, 0, len(status.Temperatures))
	for k := range status.Temperatures {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, sensor := range keys {
		builder.WriteString("📡 ")
		builder.WriteString(sensor)
		builder.WriteString(": ")
		builder.WriteString(formatTemp(status.Temperatures[sensor]))
		builder.WriteString("\n")
	}
	builder.WriteString(renderSummary(status))
	return builder.String()
}
