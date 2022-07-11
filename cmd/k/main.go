package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/kevinwylder/k/api"
	"github.com/kevinwylder/k/fs"
)

type UserSettings struct {
	DataDir  string
	TmpDir   string
	Name 	 string
	Server 	 string
}

func NewUserSettings() (*UserSettings, error) {
	var settings UserSettings
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("get home dir: %w", err)
	}
	configDir := fmt.Sprintf("%s/.config/k", home)
	if _, err := os.Stat(configDir); err != nil {
		err = os.MkdirAll(configDir, 0o755)
		if err != nil {
			return nil, fmt.Errorf("make config dir: %w", err)
		}
	}
	configFile := path.Join(configDir, "settings.json")
	f, err := os.Open(configFile)
	if err != nil {
		settings.DataDir = path.Join(configDir, "data")
		settings.TmpDir = path.Join(configDir, ".tmp")
		host, err := os.Hostname()
		if err != nil {
			host = "unknown"
		}
		settings.Name = host
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
	settings, err := NewUserSettings()
	if err != nil {
		log.Fatalf("load user settings: %v", err)
	}

	data, err := fs.NewStorageDir(settings.Name, settings.DataDir, settings.TmpDir)
	if err != nil {
		log.Fatalf("storage dir: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	t := time.Now()
	var server *api.Client

	if settings.Server != "" {
		server = api.NewClient(settings.Server, data)
		err = server.DownloadDay(ctx, t)
		if err != nil {
			log.Fatalf("Connect to server: %v", err)
		}
	}

	f, err := data.NewSegmentFile(t)
	if err != nil {
		log.Fatalf("segment file: %v", err)
	}

	cmd := exec.Command(os.Getenv("EDITOR"), f.Path)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		log.Fatalf("editor exited %v", err)
	}

	segment, err := f.Read()
	if err != nil {
		log.Fatalf("read segment: %v", err)
	}

	if server != nil {
		err = server.UploadSegment(ctx, t, segment)
		if err != nil {
			log.Fatalf("upload segment: %v", err)
		}
	} else {
		err = data.Write(t, segment, false)
		if err != nil {
			log.Fatalf("append segment: %v", err)
		}
	}
}
