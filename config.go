package main

import (
	"os"
	"path/filepath"

	toml "github.com/pelletier/go-toml/v2"
)

type Config struct {
	Channel    string `toml:"channel" json:"channel"`
	Token      string `toml:"token" json:"token"`
	SpeakerID  int    `toml:"speaker_id" json:"speaker_id"`
	ReadName   bool   `toml:"read_name" json:"read_name"`
	NameSuffix string `toml:"name_suffix" json:"name_suffix"`
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
	return filepath.Join(dir, "twitch-tts", "config.toml")
}

func (a *App) LoadConfig() Config {
	data, err := os.ReadFile(configPath())
	if err != nil {
		return defaultConfig()
	}
	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return defaultConfig()
	}
	return cfg
}

func (a *App) SaveConfig(cfg Config) error {
	p := configPath()
	if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
		return err
	}
	data, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0600)
}
