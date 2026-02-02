package app

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

func selectSeason(seasons []Season, now time.Time) Season {
	if len(seasons) == 0 {
		return Season{}
	}
	nowMinutes := now.Hour()*60 + now.Minute()
	for _, season := range seasons {
		start, err := parseTime(season.TimeStart)
		if err != nil {
			continue
		}
		end, err := parseTime(season.TimeEnd)
		if err != nil {
			continue
		}
		if withinWindow(nowMinutes, start, end) {
			return season
		}
	}
	return seasons[0]
}

func parseTime(value string) (int, error) {
	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return 0, errors.New("invalid time")
	}
	hour, err := time.Parse("15:04", value)
	if err != nil {
		return 0, err
	}
	return hour.Hour()*60 + hour.Minute(), nil
}

func withinWindow(nowMinutes, start, end int) bool {
	if start == end {
		return true
	}
	if start < end {
		return nowMinutes >= start && nowMinutes < end
	}
	return nowMinutes >= start || nowMinutes < end
}

func average(values []float64) float64 {
	if len(values) == 0 {
		return math.NaN()
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func formatTemp(temp float64) string {
	return fmt.Sprintf("%.1f°C", temp)
}

func formatBattery(battery float64) string {
	return fmt.Sprintf("%.0f%%", battery)
}

func timeTickerMinute() *time.Ticker {
	return time.NewTicker(time.Minute)
}
