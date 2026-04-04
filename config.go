package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Channel    string `json:"channel"`
	Token      string `json:"token"`
	SpeakerID  int    `json:"speaker_id"`
	ReadName   bool   `json:"read_name"`
	NameSuffix string `json:"name_suffix"`
}

func defaultConfig() Config {
	return Config{SpeakerID: 1, ReadName: true, NameSuffix: "さん"}
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
		return defaultConfig()
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return defaultConfig()
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
