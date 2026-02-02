package app

import (
	"context"
	"time"
)

func (a *App) RunBatteryChecks(ctxDone <-chan struct{}) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ctxDone:
			return
		case <-ticker.C:
			a.checkBatteries()
		}
	}
}

func (a *App) checkBatteries() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for _, room := range a.cfg.Rooms {
		for _, sensor := range room.Sensors {
			battery, ok := a.readSensorBattery(ctx, sensor)
			if ok && battery < 25 {
				a.broadcast(formatBatteryAlert(room.Name, sensor, battery))
			}
		}
		if room.Thermostat != "" {
			battery, ok := a.readThermostatBattery(ctx, room.Thermostat)
			if ok && battery < 25 {
				a.broadcast(formatBatteryAlert(room.Name, room.Thermostat, battery))
			}
		}
	}
}

func formatBatteryAlert(roomName, deviceID string, battery float64) string {
	return "🔋 " + roomName + " (" + deviceID + "): батарея " + formatBattery(battery)
}
