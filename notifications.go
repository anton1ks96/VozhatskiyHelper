package main

import (
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func StartNotificationScheduler(bot *tgbotapi.BotAPI, chatIDs []int64) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		for _, chatID := range chatIDs {
			CheckAndSendNotifications(bot, chatID)
		}
	}
}

func CheckAndSendNotifications(bot *tgbotapi.BotAPI, chatID int64) {
	loc := time.FixedZone("GMT+5", 5*3600)
	now := time.Now().In(loc)

	for i := 0; i < len(eventsToday); i++ {
		remaining := eventsToday[i].StartTime.Sub(now)
		if remaining <= 30*time.Minute && remaining > 29*time.Minute && !eventsToday[i].notified30 {
			text := fmt.Sprintf("Через 30 минут событие \"%s\" в локации \"%s\"", eventsToday[i].Name, eventsToday[i].Location)
			msg := tgbotapi.NewMessage(chatID, text)
			if _, err := bot.Send(msg); err != nil {
			} else {
				eventsToday[i].notified30 = true
			}
		}

		if remaining <= 10*time.Minute && remaining > 9*time.Minute && !eventsToday[i].notified10 {
			text := fmt.Sprintf("Скоро начинается \"%s\" – пора собираться! Локация - \"%s\"", eventsToday[i].Name, eventsToday[i].Location)
			msg := tgbotapi.NewMessage(chatID, text)
			if _, err := bot.Send(msg); err != nil {
			} else {
				eventsToday[i].notified10 = true
			}
		}
	}

	for i := 0; i < len(eventsTomorrow); i++ {
		remaining := eventsTomorrow[i].StartTime.Sub(now)
		if remaining <= 30*time.Minute && remaining > 29*time.Minute && !eventsTomorrow[i].notified30 {
			text := fmt.Sprintf("Через 30 минут событие \"%s\" в локации \"%s\"", eventsTomorrow[i].Name, eventsTomorrow[i].Location)
			msg := tgbotapi.NewMessage(chatID, text)
			if _, err := bot.Send(msg); err != nil {
			} else {
				eventsTomorrow[i].notified30 = true
			}
		}

		if remaining <= 10*time.Minute && remaining > 9*time.Minute && !eventsTomorrow[i].notified10 {
			text := fmt.Sprintf("Скоро начинается \"%s\" – пора собираться! Локация - \"%s\"", eventsTomorrow[i].Name, eventsTomorrow[i].Location)
			msg := tgbotapi.NewMessage(chatID, text)
			if _, err := bot.Send(msg); err != nil {
			} else {
				eventsTomorrow[i].notified10 = true
			}
		}
	}
}
