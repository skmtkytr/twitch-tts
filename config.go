package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Channel   string `json:"channel"`
	Token     string `json:"token"`
	SpeakerID int    `json:"speaker_id"`
}

func configPath() string {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, "twitch-tts", "config.json")
}

func (a *App) LoadConfig() Config {
	data, err := os.ReadFile(configPath())
	if err != nil {
		return Config{SpeakerID: 1}
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{SpeakerID: 1}
	}
	return cfg
}

func (a *App) SaveConfig(cfg Config) error {
	p := configPath()
	if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0600)
}
