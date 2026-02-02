package app

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/redis/go-redis/v9"
)

const (
	sensorPrefix     = "sensor:"
	thermostatPrefix = "thermostat:"
	valkeyTTL        = 24 * time.Hour
	manualTTL        = time.Hour
	manualPrefix     = "manual:"
)

type App struct {
	cfg          Config
	valkey       *redis.Client
	mqttClient   mqtt.Client
	bot          *tgbotapi.BotAPI
	chatMu       sync.RWMutex
	chatIDs      map[int64]struct{}
	deviceToRoom map[string]string
}

func New(cfg Config) (*App, error) {
	app := &App{cfg: cfg, chatIDs: make(map[int64]struct{}), deviceToRoom: make(map[string]string)}
	if err := app.initValkey(); err != nil {
		return nil, fmt.Errorf("valkey: %w", err)
	}
	if err := app.initMQTT(); err != nil {
		return nil, fmt.Errorf("mqtt: %w", err)
	}
	if err := app.initBot(); err != nil {
		return nil, fmt.Errorf("telegram: %w", err)
	}
	app.buildDeviceMap()
	return app, nil
}

func (a *App) initValkey() error {
	opt, err := redis.ParseURL(a.cfg.ValkeyURL)
	if err != nil {
		return err
	}
	client := redis.NewClient(opt)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return err
	}
	a.valkey = client
	return nil
}

func (a *App) initMQTT() error {
	opts := mqtt.NewClientOptions().AddBroker(a.cfg.URLMQTT)
	if a.cfg.UserMQTT != "" {
		opts.SetUsername(a.cfg.UserMQTT)
		opts.SetPassword(a.cfg.PassMQTT)
	}
	opts.SetDefaultPublishHandler(a.handleMQTT)
	opts.OnConnectionLost = func(_ mqtt.Client, err error) {
		log.Printf("mqtt connection lost: %v", err)
	}
	opts.OnConnect = func(client mqtt.Client) {
		for _, room := range a.cfg.Rooms {
			for _, sensor := range room.Sensors {
				topic := fmt.Sprintf("zigbee2mqtt/%s", sensor)
				if token := client.Subscribe(topic, 0, a.handleMQTT); token.Wait() && token.Error() != nil {
					log.Printf("subscribe %s: %v", topic, token.Error())
				}
			}
			if room.Thermostat != "" {
				topic := fmt.Sprintf("zigbee2mqtt/%s", room.Thermostat)
				if token := client.Subscribe(topic, 0, a.handleMQTT); token.Wait() && token.Error() != nil {
					log.Printf("subscribe %s: %v", topic, token.Error())
				}
			}
		}
		log.Printf("mqtt subscriptions ready")
	}
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	a.mqttClient = client
	return nil
}

func (a *App) initBot() error {
	bot, err := tgbotapi.NewBotAPI(a.cfg.TGToken)
	if err != nil {
		return err
	}
	bot.Debug = false
	a.bot = bot
	return nil
}

func (a *App) buildDeviceMap() {
	for _, room := range a.cfg.Rooms {
		for _, sensor := range room.Sensors {
			a.deviceToRoom[sensor] = room.Name
		}
		if room.Thermostat != "" {
			a.deviceToRoom[room.Thermostat] = room.Name
		}
	}
}

func (a *App) Shutdown() {
	if a.mqttClient != nil {
		a.mqttClient.Disconnect(250)
	}
	if a.valkey != nil {
		_ = a.valkey.Close()
	}
}
