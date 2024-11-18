package config

import (
    "encoding/json"
    "os"
    "path/filepath"
)

const configFileName = ".gatorconfig.json"
const defaultDBURL = "postgres://danielmariathasan@localhost:5432/gator?sslmode=disable"

type Config struct {
    CurrentUserName string `json:"current_user_name"`
    DatabaseURL     string `json:"database_url"`
}

func (cfg *Config) SetUser(username string) error {
    cfg.CurrentUserName = username
    return Write(*cfg)
}

func Read() (Config, error) {
    fullPath, err := getConfigFilePath()
    if err != nil {
        return Config{}, err
    }

    file, err := os.Open(fullPath)
    if os.IsNotExist(err) {
        // File doesn't exist, create new config with defaults
        cfg := New()
        err = Write(cfg)
        if err != nil {
            return Config{}, err
        }
        return cfg, nil
    } else if err != nil {
        return Config{}, err
    }
    defer file.Close()

    decoder := json.NewDecoder(file)
    cfg := Config{}
    err = decoder.Decode(&cfg)
    if err != nil {
        return Config{}, err
    }

    return cfg, nil
}

func getConfigFilePath() (string, error) {
    home, err := os.UserHomeDir()
    if err != nil {
        return "", err
    }
    fullPath := filepath.Join(home, configFileName)
    return fullPath, nil
}

func Write(cfg Config) error {
    fullPath, err := getConfigFilePath()
    if err != nil {
        return err
    }

    file, err := os.Create(fullPath)
    if err != nil {
        return err
    }
    defer file.Close()

    encoder := json.NewEncoder(file)
    err = encoder.Encode(cfg)
    if err != nil {
        return err
    }

    return nil
}

func New() Config {
    return Config{
        DatabaseURL: defaultDBURL,
    }
}
