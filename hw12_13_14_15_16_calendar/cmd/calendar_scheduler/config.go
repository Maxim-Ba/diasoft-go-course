package main

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Logger   LoggerConf   `mapstructure:"logger"`
	Database DatabaseConf `mapstructure:"database"`
	Kafka    KafkaConf    `mapstructure:"kafka"`
	Storage  StorageConf  `mapstructure:"storage"`
	Schedule ScheduleConf `mapstructure:"schedule"`
}

type LoggerConf struct {
	Level string `mapstructure:"level"`
}

type DatabaseConf struct {
	DSN string `mapstructure:"dsn"`
}

type KafkaConf struct {
	Brokers []string `mapstructure:"brokers"`
	Topic   string   `mapstructure:"topic"`
}

type StorageConf struct {
	Type string `mapstructure:"type"`
}

type ScheduleConf struct {
	ScanInterval string `mapstructure:"scan_interval"`
}

func NewConfig(configPath string) (*Config, error) {
	v := viper.New()

	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}
