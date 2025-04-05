package main

import (
	"fmt"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func ScheduleNotifications(bot *tgbotapi.BotAPI, chatID int64) {
	now := time.Now()

	var allEvents []Event
	allEvents = append(allEvents, eventsToday...)
	allEvents = append(allEvents, eventsTomorrow...)

	for _, event := range allEvents {
		reminder30 := event.StartTime.Add(-30 * time.Minute)
		reminder10 := event.StartTime.Add(-10 * time.Minute)

		if reminder30.After(now) {
			go func(ev Event, t time.Time) {
				time.Sleep(time.Until(t))
				text := fmt.Sprintf("Через 30 минут событие \"%s\" в локации \"%s\"", ev.Name, ev.Location)
				msg := tgbotapi.NewMessage(chatID, text)
				_, err := bot.Send(msg)
				if err != nil {
					log.Println("Ошибка отправки уведомления за 30 мин для события", ev.Name, ":", err)
				} else {
					log.Printf("Уведомление за 30 мин для события '%s' успешно отправлено в чат %d", ev.Name, chatID)
				}
			}(event, reminder30)
		}

		if reminder10.After(now) {
			go func(ev Event, t time.Time) {
				time.Sleep(time.Until(t))
				text := fmt.Sprintf("Скоро начинается \"%s\" – пора собираться! Локация - \"%s\"", ev.Name, ev.Location)
				msg := tgbotapi.NewMessage(chatID, text)
				_, err := bot.Send(msg)
				if err != nil {
					log.Println("Ошибка отправки уведомления за 10 мин для события", ev.Name, ":", err)
				} else {
					log.Printf("Уведомление за 10 мин для события '%s' успешно отправлено в чат %d", ev.Name, chatID)
				}
			}(event, reminder10)
		}
	}
}
