package app

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

func (a *App) collectStatuses() []RoomStatus {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	statuses := make([]RoomStatus, 0, len(a.cfg.Rooms))
	for _, room := range a.cfg.Rooms {
		status := RoomStatus{
			Room:         room,
			Temperatures: make(map[string]float64),
			ObservedAt:   time.Now(),
		}
		var temps []float64
		for _, sensor := range room.Sensors {
			value, err := a.valkey.Get(ctx, fmt.Sprintf("%s%s:last", sensorPrefix, sensor)).Result()
			if err != nil {
				continue
			}
			var payload SensorPayload
			if err := json.Unmarshal([]byte(value), &payload); err != nil {
				continue
			}
			status.Temperatures[sensor] = payload.Temperature
			temps = append(temps, payload.Temperature)
		}
		status.AverageTemp = average(temps)
		status.Season = selectSeason(room.Seasons, status.ObservedAt)
		if room.Thermostat != "" {
			value, err := a.valkey.Get(ctx, fmt.Sprintf("%s%s:last", thermostatPrefix, room.Thermostat)).Result()
			if err == nil {
				var payload ThermostatPayload
				if err := json.Unmarshal([]byte(value), &payload); err == nil {
					status.ThermostatSetpoint = payload.CurrentHeatingSetpoint
					status.ThermostatOn = payload.CurrentHeatingSetpoint >= 20
				}
			}
		}
		statuses = append(statuses, status)
	}
	return statuses
}

func (a *App) readSensorBattery(ctx context.Context, sensorID string) (float64, bool) {
	value, err := a.valkey.Get(ctx, fmt.Sprintf("%s%s:last", sensorPrefix, sensorID)).Result()
	if err != nil {
		return 0, false
	}
	var payload SensorPayload
	if err := json.Unmarshal([]byte(value), &payload); err != nil {
		return 0, false
	}
	return payload.Battery, true
}

func (a *App) readThermostatBattery(ctx context.Context, deviceID string) (float64, bool) {
	value, err := a.valkey.Get(ctx, fmt.Sprintf("%s%s:last", thermostatPrefix, deviceID)).Result()
	if err != nil {
		return 0, false
	}
	var payload ThermostatPayload
	if err := json.Unmarshal([]byte(value), &payload); err != nil {
		return 0, false
	}
	return payload.Battery, true
}

func renderSummary(status RoomStatus) string {
	builder := &strings.Builder{}
	builder.WriteString(renderAverageLine(status))
	builder.WriteString(renderSeasonLine(status))
	if status.Room.Thermostat != "" {
		if status.ThermostatOn {
			builder.WriteString("🔥 Идет обогрев")
		} else {
			builder.WriteString("🧊 Обогрев выключен")
		}
	}
	return builder.String()
}

func renderAverageLine(status RoomStatus) string {
	if math.IsNaN(status.AverageTemp) {
		return "📊 Средняя температура: нет данных\n"
	}
	return "📊 Средняя температура: " + formatTemp(status.AverageTemp) + " | Время: " + status.ObservedAt.Format("15:04") + "\n"
}

func renderSeasonLine(status RoomStatus) string {
	if status.Season.Name == "" {
		return ""
	}
	return "⏰ Временной слот: " + status.Season.Name + "\n"
}

func renderSensors(status RoomStatus) string {
	keys := make([]string, 0, len(status.Temperatures))
	for k := range status.Temperatures {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	builder := &strings.Builder{}
	for _, sensor := range keys {
		builder.WriteString("📡 ")
		builder.WriteString(sensor)
		builder.WriteString(": ")
		builder.WriteString(formatTemp(status.Temperatures[sensor]))
		builder.WriteString("\n")
	}
	return builder.String()
}
