package app

import (
	"context"
	"fmt"
	"time"
)

func (a *App) storeManualOverride(roomID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	key := fmt.Sprintf("%s%s", manualPrefix, roomID)
	_ = a.valkey.Set(ctx, key, time.Now().UTC().Format(time.RFC3339), manualTTL).Err()
}

func (a *App) hasManualOverride(roomID string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	key := fmt.Sprintf("%s%s", manualPrefix, roomID)
	_, err := a.valkey.Get(ctx, key).Result()
	return err == nil
}
