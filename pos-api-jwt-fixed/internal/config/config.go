package config

import (
    "os"
    "gopkg.in/yaml.v3"
)

type Config struct {
    Server struct {
        Port string `yaml:"port"`
        Env  string `yaml:"env"`
    } `yaml:"server"`

    Database struct {
        Host     string `yaml:"host"`
        Port     string `yaml:"port"`
        User     string `yaml:"user"`
        Password string `yaml:"password"`
        Name     string `yaml:"name"`
    } `yaml:"database"`
}

var Cfg Config

// 讀取 YAML 設定檔
func LoadConfig(path string) error {
    file, err := os.ReadFile(path)
    if err != nil {
        return err
    }

    err = yaml.Unmarshal(file, &Cfg)
    if err != nil {
        return err
    }

    return nil
}

