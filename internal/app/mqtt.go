package app

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func (a *App) handleMQTT(_ mqtt.Client, msg mqtt.Message) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	topicParts := strings.Split(msg.Topic(), "/")
	if len(topicParts) < 2 {
		return
	}
	deviceID := topicParts[len(topicParts)-1]
	payload := msg.Payload()

	for _, room := range a.cfg.Rooms {
		for _, sensor := range room.Sensors {
			if sensor == deviceID {
				key := fmt.Sprintf("%s%s:last", sensorPrefix, deviceID)
				_ = a.valkey.Set(ctx, key, payload, valkeyTTL).Err()
				return
			}
		}
		if room.Thermostat == deviceID {
			key := fmt.Sprintf("%s%s:last", thermostatPrefix, deviceID)
			_ = a.valkey.Set(ctx, key, payload, valkeyTTL).Err()
			return
		}
	}
}

func (a *App) publishThermostat(room Room, turnOn bool) error {
	if room.Thermostat == "" {
		return nil
	}
	payload := map[string]any{
		"current_heating_setpoint": setpointFor(turnOn),
	}
	if room.ThermoAdv {
		if turnOn {
			payload["preset"] = "on"
		} else {
			payload["preset"] = "off"
		}
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	topic := fmt.Sprintf("zigbee2mqtt/%s/set", room.Thermostat)
	token := a.mqttClient.Publish(topic, 0, false, data)
	token.Wait()
	return token.Error()
}

func setpointFor(turnOn bool) int {
	if turnOn {
		return 30
	}
	return 5
}
