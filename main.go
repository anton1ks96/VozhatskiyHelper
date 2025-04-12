package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

var pendingSchedulePath string

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Ошибка загрузки .env файла:", err)
	}
	token := os.Getenv("TELEGRAM_TOKEN")

	adminChatIDsStr := os.Getenv("ADMIN_CHAT_IDS")
	adminChatIDs, err := parseChatIDs(adminChatIDsStr)
	if err != nil {
		log.Fatalf("Ошибка парсинга ADMIN_CHAT_IDS: %v", err)
	}

	notificationChatIDsStr := os.Getenv("NOTIFICATION_CHAT_IDS")
	notificationChatIDs, err := parseChatIDs(notificationChatIDsStr)
	if err != nil {
		log.Fatalf("Ошибка парсинга NOTIFICATION_CHAT_IDS: %v", err)
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = true
	log.Printf("Авторизован как %s", bot.Self.UserName)
	go StartScheduleUpdater()
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			if update.Message.Document != nil && contains(adminChatIDs, update.Message.Chat.ID) {
				fileID := update.Message.Document.FileID
				file, err := bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
				if err != nil {
					log.Println("Ошибка получения файла:", err)
					continue
				}
				fileURL := file.Link(token)

				resp, err := http.Get(fileURL)
				if err != nil {
					log.Println("Ошибка загрузки файла:", err)
					continue
				}
				defer resp.Body.Close()
				tempFile, err := os.CreateTemp("", "schedule-*.xlsx")
				if err != nil {
					log.Println("Ошибка создания временного файла:", err)
					continue
				}
				pendingSchedulePath = tempFile.Name()
				_, err = io.Copy(tempFile, resp.Body)
				if err != nil {
					log.Println("Ошибка сохранения файла:", err)
					continue
				}
				tempFile.Close()

				keyboard := tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Обновить на сегодня", "update_today"),
						tgbotapi.NewInlineKeyboardButtonData("Добавить на завтра", "update_tomorrow"),
					),
				)

				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выберите действие для загруженного расписания:")
				msg.ReplyMarkup = keyboard
				bot.Send(msg)
			}
		}

		if update.CallbackQuery != nil {
			callbackData := update.CallbackQuery.Data
			var dayOffset int
			if callbackData == "update_today" {
				dayOffset = 0
			} else if callbackData == "update_tomorrow" {
				dayOffset = 1
			} else {
				callback := tgbotapi.CallbackConfig{
					CallbackQueryID: update.CallbackQuery.ID,
					Text:            "Неизвестное действие",
				}
				_, err = bot.Request(callback)
				if err != nil {
					log.Println("Ошибка ответа на callback:", err)
				}
				continue
			}

			loadedEvents, err := LoadSchedule(pendingSchedulePath, dayOffset)
			if err != nil {
				log.Println("Ошибка загрузки расписания:", err)
				callback := tgbotapi.CallbackConfig{
					CallbackQueryID: update.CallbackQuery.ID,
					Text:            "Ошибка загрузки расписания",
				}
				_, err = bot.Request(callback)
				if err != nil {
					log.Println("Ошибка ответа на callback:", err)
				}
				continue
			}

			SetEvents(loadedEvents, dayOffset)

			go StartNotificationScheduler(bot, notificationChatIDs)

			var dayText string
			switch dayOffset {
			case 0:
				dayText = "на сегодня"
			case 1:
				dayText = "на завтра"
			default:
				dayText = fmt.Sprintf("на %d день вперёд", dayOffset)
			}

			scheduleMsgText := fmt.Sprintf("📅 Расписание успешно обновлено %s!\n\n", dayText)
			for _, ev := range loadedEvents {
				scheduleMsgText += fmt.Sprintf("• %s | %s | %s\n", ev.Name, ev.Location, ev.StartTime.Format("15:04"))
			}

			for _, chatID := range notificationChatIDs {
				msg := tgbotapi.NewMessage(chatID, scheduleMsgText)
				if _, err := bot.Send(msg); err != nil {
					log.Printf("Ошибка отправки сообщения в чат %d: %v", chatID, err)
				}
			}

			editMsg := tgbotapi.NewEditMessageText(
				update.CallbackQuery.Message.Chat.ID,
				update.CallbackQuery.Message.MessageID,
				scheduleMsgText,
			)
			_, err = bot.Send(editMsg)
			if err != nil {
				log.Println("Ошибка редактирования сообщения:", err)
			}

			callback := tgbotapi.CallbackConfig{
				CallbackQueryID: update.CallbackQuery.ID,
				Text:            "Расписание обновлено",
			}
			_, err = bot.Request(callback)
			if err != nil {
				log.Println("Ошибка ответа на callback:", err)
			}

			os.Remove(pendingSchedulePath)
			pendingSchedulePath = ""
		}

		if update.Message != nil && update.Message.IsCommand() {
			switch update.Message.Command() {
			case "next":
				reply := GetNextEvents()
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
				bot.Send(msg)
			case "schedule":
				scheduleText := GetFullSchedule()
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, scheduleText)
				_, err := bot.Send(msg)
				if err != nil {
					log.Println("Ошибка отправки расписания:", err)
				}
			}
		}
	}
}

func parseChatIDs(idsStr string) ([]int64, error) {
	var ids []int64
	parts := strings.Split(idsStr, ",")
	for _, part := range parts {
		id, err := strconv.ParseInt(strings.TrimSpace(part), 10, 64)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func contains(slice []int64, id int64) bool {
	for _, v := range slice {
		if v == id {
			return true
		}
	}
	return false
}
