package app

import "time"

type SensorPayload struct {
	Battery     float64 `json:"battery"`
	Humidity    any     `json:"humidity"`
	Temperature float64 `json:"temperature"`
}

type ThermostatPayload struct {
	Battery                float64 `json:"battery"`
	CurrentHeatingSetpoint float64 `json:"current_heating_setpoint"`
	Preset                 string  `json:"preset"`
}

type RoomStatus struct {
	Room               Room
	AverageTemp        float64
	Temperatures       map[string]float64
	Season             Season
	ThermostatOn       bool
	ThermostatSetpoint float64
	ObservedAt         time.Time
}
