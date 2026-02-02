package app

import (
	"log"
	"math"
	"strings"
)

func (a *App) RunRuleLoop(ctxDone <-chan struct{}) {
	ticker := timeTickerMinute()
	defer ticker.Stop()
	for {
		select {
		case <-ctxDone:
			return
		case <-ticker.C:
			for _, status := range a.collectStatuses() {
				if status.Room.Thermostat == "" {
					continue
				}
				if a.hasManualOverride(status.Room.ID) {
					continue
				}
				a.applyRules(status)
			}
		}
	}
}

func (a *App) applyRules(status RoomStatus) {
	if math.IsNaN(status.AverageTemp) {
		return
	}
	if len(status.Room.Seasons) == 0 {
		if status.ThermostatOn {
			if err := a.publishThermostat(status.Room, false); err != nil {
				log.Printf("turn off: %v", err)
				return
			}
			a.broadcast(renderActionMessage(status, "выключение", "нет правил"))
		}
		return
	}
	min := status.Season.Min
	max := status.Season.Max
	if status.AverageTemp < min && !status.ThermostatOn {
		if err := a.publishThermostat(status.Room, true); err != nil {
			log.Printf("turn on: %v", err)
			return
		}
		a.broadcast(renderActionMessage(status, "включение", "холодно"))
		return
	}
	if status.AverageTemp > max && status.ThermostatOn {
		if err := a.publishThermostat(status.Room, false); err != nil {
			log.Printf("turn off: %v", err)
			return
		}
		a.broadcast(renderActionMessage(status, "выключение", "жарко"))
	}
}

func renderActionMessage(status RoomStatus, action, reason string) string {
	builder := &strings.Builder{}
	builder.WriteString(renderSensors(status))
	builder.WriteString(renderAverageLine(status))
	builder.WriteString(renderSeasonLine(status))
	builder.WriteString("🔥 ")
	builder.WriteString(status.Room.Name)
	builder.WriteString(": ")
	builder.WriteString(action)
	builder.WriteString(" | ")
	builder.WriteString(reason)
	return builder.String()
}
