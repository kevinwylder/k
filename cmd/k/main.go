package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/kevinwylder/k/api"
	"github.com/kevinwylder/k/files"
)

type UserSettings struct {
	ServerAddr string
	CacheDir   string
}

func LoadFromSettings() (*UserSettings, error) {
	var settings UserSettings
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("get config dir: %w", err)
	}
	configDir = path.Join(configDir, "k")
	if _, err := os.Stat(configDir); err != nil {
		err = os.MkdirAll(configDir, 0o755)
		if err != nil {
			return nil, fmt.Errorf("make config dir: %w", err)
		}
	}
	configFile := path.Join(configDir, "settings.json")
	f, err := os.Open(configFile)
	if err != nil {
		settings.CacheDir = path.Join(configDir, "cache")
		settings.ServerAddr = "127.0.0.1"
		f, err := os.Create(configFile)
		if err != nil {
			return nil, fmt.Errorf("create file: %w", err)
		}
		err = json.NewEncoder(f).Encode(&settings)
		if err != nil {
			return nil, fmt.Errorf("write file: %w", err)
		}
		return &settings, nil
	}
	err = json.NewDecoder(f).Decode(&settings)
	if err != nil {
		return nil, fmt.Errorf("parse settings json %s: %w", configFile, err)
	}
	return &settings, nil
}

func main() {
	settings, err := LoadFromSettings()
	if err != nil {
		log.Fatalf("load user settings: %v", err)
	}
	cache := files.NewCacheDir(settings.CacheDir)
	client := api.NewClient(settings.ServerAddr, cache)
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	err = client.Check(ctx)
	if err != nil {
		log.Fatalf("Connect to server: %v", err)
	}
}
