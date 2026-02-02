package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"smarthome/internal/app"
)

func main() {
	configPath := flag.String("config", "config.json", "path to config JSON")
	flag.Parse()

	cfg, err := app.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	controller, err := app.New(cfg)
	if err != nil {
		log.Fatalf("init app: %v", err)
	}
	defer controller.Shutdown()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go controller.RunBot(ctx)
	go controller.RunRuleLoop(ctx)
	go controller.RunBatteryChecks(ctx)

	<-ctx.Done()
}
