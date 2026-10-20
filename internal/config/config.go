package config

import (
	"encoding/json"
	"os"
)

type PackData struct {
	Repo           string   `json:"repo"`
	Version        string   `json:"version"`
	ExtractedFiles []string `json:"extracted_files"`
}

type AppConfig struct {
	GamePath       string              `json:"game_path"`
	AppLang        string              `json:"app_lang"`
	InstalledPacks map[string]PackData `json:"installed_packs"`
}

const configFile = "config.json"

func Load() (*AppConfig, error) {
	cfg := &AppConfig{
		AppLang:        "en",
		InstalledPacks: make(map[string]PackData),
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	err = json.Unmarshal(data, cfg)
	if cfg.AppLang == "" {
		cfg.AppLang = "en"
	}
	return cfg, err
}

func (c *AppConfig) Save() error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configFile, data, 0644)
}
