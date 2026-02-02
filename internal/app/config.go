package app

import (
	"encoding/json"
	"errors"
	"os"
)

type Config struct {
	TGToken   string `json:"tg_token"`
	URLMQTT   string `json:"url_mqtt"`
	UserMQTT  string `json:"user_mqtt"`
	PassMQTT  string `json:"pass_mqtt"`
	ValkeyURL string `json:"valkey_url"`
	ChatID    int64  `json:"chat_id"`
	Rooms     []Room `json:"rooms"`
}

type Room struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Sensors    []string `json:"sensors"`
	Thermostat string   `json:"thermostat"`
	ThermoAdv  bool     `json:"thermo_adv"`
	Seasons    []Season `json:"seasons"`
}

type Season struct {
	TimeStart string  `json:"time_start"`
	TimeEnd   string  `json:"time_end"`
	Min       float64 `json:"min"`
	Max       float64 `json:"max"`
	Name      string  `json:"name"`
}

func LoadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	if cfg.TGToken == "" || cfg.URLMQTT == "" || cfg.ValkeyURL == "" {
		return Config{}, errors.New("missing required configuration values")
	}
	return cfg, nil
}
