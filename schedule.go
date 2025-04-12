package main

import (
	"fmt"
	"log"
	"time"

	"github.com/xuri/excelize/v2"
)

type Event struct {
	Name       string
	Location   string
	StartTime  time.Time
	notified30 bool
	notified10 bool
}

var eventsToday []Event
var eventsTomorrow []Event

func LoadSchedule(filePath string, dayOffset int) ([]Event, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, err
	}
	sheetName := f.GetSheetName(0)
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, err
	}

	var schedule []Event
	loc := time.FixedZone("GMT+5", 5*3600)
	targetDate := time.Now().In(loc).Add(time.Duration(dayOffset) * 24 * time.Hour).Format("2006-01-02")
	for i, row := range rows {
		if i == 0 {
			continue
		}
		if len(row) < 3 {
			continue
		}
		name := row[0]
		location := row[1]
		timeStr := row[2] // ожидается формат "08:30:00"
		fullTimeStr := fmt.Sprintf("%s %s", targetDate, timeStr)
		t, err := time.ParseInLocation("2006-01-02 15:04:05", fullTimeStr, loc)
		if err != nil {
			log.Println("Ошибка парсинга времени для события", name, ":", err)
			continue
		}
		reminder30 := t.Add(-30 * time.Minute)
		reminder10 := t.Add(-10 * time.Minute)
		log.Printf("Событие '%s' (место: %s) запланировано на %s. Уведомления: за 30 мин (%s), за 10 мин (%s)",
			name, location, t.Format("15:04:05"), reminder30.Format("15:04:05"), reminder10.Format("15:04:05"))

		schedule = append(schedule, Event{
			Name:      name,
			Location:  location,
			StartTime: t,
		})
	}
	return schedule, nil
}

func SetEvents(newEvents []Event, dayOffset int) {
	if dayOffset == 0 {
		eventsToday = newEvents
	} else if dayOffset == 1 {
		eventsTomorrow = newEvents
	}
}

func StartScheduleUpdater() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		ShiftScheduleIfNeeded()
		RemovePastEvents()
	}
}
