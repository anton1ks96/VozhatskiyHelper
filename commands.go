package main

import (
	"fmt"
	"sort"
	"time"
)

func GetNextEvents() string {
	RemovePastEvents()

	loc := time.FixedZone("GMT+5", 5*3600)
	now := time.Now().In(loc)

	var allEvents []Event
	allEvents = append(allEvents, eventsToday...)
	allEvents = append(allEvents, eventsTomorrow...)

	var upcoming []Event
	for _, ev := range allEvents {
		if ev.StartTime.After(now) {
			upcoming = append(upcoming, ev)
		}
	}

	if len(upcoming) == 0 {
		return "Пока что все события завершены. Хорошего дня!"
	}

	sort.Slice(upcoming, func(i, j int) bool {
		return upcoming[i].StartTime.Before(upcoming[j].StartTime)
	})

	count := 2
	if len(upcoming) < 2 {
		count = len(upcoming)
	}

	response := "Следующие события:\n"
	for i := 0; i < count; i++ {
		event := upcoming[i]
		var dayLabel string
		if event.StartTime.In(loc).Format("2006-01-02") == now.Format("2006-01-02") {
			dayLabel = "Сегодня"
		} else if event.StartTime.In(loc).Format("2006-01-02") == now.Add(24*time.Hour).Format("2006-01-02") {
			dayLabel = "Завтра"
		}
		response += fmt.Sprintf("%s: \"%s\" в %s\n", dayLabel, event.Name, event.StartTime.In(loc).Format("15:04"))
	}
	return response
}

func RemovePastEvents() {
	now := time.Now()
	var filteredToday []Event
	for _, ev := range eventsToday {
		if ev.StartTime.After(now) {
			filteredToday = append(filteredToday, ev)
		}
	}
	eventsToday = filteredToday

	var filteredTomorrow []Event
	for _, ev := range eventsTomorrow {
		if ev.StartTime.After(now) {
			filteredTomorrow = append(filteredTomorrow, ev)
		}
	}
	eventsTomorrow = filteredTomorrow
}

func GetFullSchedule() string {
	loc := time.FixedZone("GMT+5", 5*3600)
	response := "Полное расписание:\n\n"

	if len(eventsToday) > 0 {
		response += "Сегодня:\n"
		sort.Slice(eventsToday, func(i, j int) bool {
			return eventsToday[i].StartTime.Before(eventsToday[j].StartTime)
		})
		for _, ev := range eventsToday {
			response += fmt.Sprintf("• %s в %s, локация: %s\n", ev.Name, ev.StartTime.In(loc).Format("15:04"), ev.Location)
		}
		response += "\n"
	} else {
		response += "Сегодня расписания нет.\n\n"
	}

	if len(eventsTomorrow) > 0 {
		response += "Завтра:\n"
		sort.Slice(eventsTomorrow, func(i, j int) bool {
			return eventsTomorrow[i].StartTime.Before(eventsTomorrow[j].StartTime)
		})
		for _, ev := range eventsTomorrow {
			response += fmt.Sprintf("• %s в %s, локация: %s\n", ev.Name, ev.StartTime.In(loc).Format("15:04"), ev.Location)
		}
	} else {
		response += "Завтра расписания нет.\n"
	}

	return response
}

func ShiftScheduleIfNeeded() {
	loc := time.FixedZone("GMT+5", 5*3600)
	todayDate := time.Now().In(loc).Format("2006-01-02")

	if len(eventsToday) > 0 {
		eventDate := eventsToday[0].StartTime.In(loc).Format("2006-01-02")
		if eventDate != todayDate {
			eventsToday = eventsTomorrow
			eventsTomorrow = nil
		}
	} else if len(eventsTomorrow) > 0 {
		eventDate := eventsTomorrow[0].StartTime.In(loc).Format("2006-01-02")
		if eventDate == todayDate {
			eventsToday = eventsTomorrow
			eventsTomorrow = nil
		}
	}
}
