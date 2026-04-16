package main

import (
	"context"
	"flag"
	"os/signal"
	"syscall"

	"github.com/Maxim-Ba/hw12_13_14_15_calendar/internal/logger"
	"github.com/Maxim-Ba/hw12_13_14_15_calendar/internal/queue"
	sqlstorage "github.com/Maxim-Ba/hw12_13_14_15_calendar/internal/storage/sql"
	"github.com/Maxim-Ba/hw12_13_14_15_calendar/internal/storer"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "configs/storer_config.yaml", "Path to configuration file")
}

func main() {
	flag.Parse()

	config, err := NewConfig(configFile)
	if err != nil {
		logger.New("ERROR").Error("failed to load config: " + err.Error())
		panic(err.Error())
	}

	logg := logger.New(config.Logger.Level)
	logg.Info("calendar storer starting...")

	storage, err := sqlstorage.New(config.Database.DSN)
	if err != nil {
		logg.Error("failed to create SQL storage: " + err.Error())
		panic(err.Error())
	}
	defer func() {
		if err := storage.Close(); err != nil {
			logg.Error("failed to close SQL storage: " + err.Error())
		}
	}()

	consumer, err := queue.NewKafkaConsumer(config.Kafka.Brokers, config.Kafka.GroupID)
	if err != nil {
		logg.Error("failed to create Kafka consumer: " + err.Error())
		if closeErr := storage.Close(); closeErr != nil {
			logg.Error("failed to close SQL storage: " + closeErr.Error())
		}
		panic(err.Error())
	}
	defer func() {
		if err := consumer.Close(); err != nil {
			logg.Error("failed to close Kafka consumer: " + err.Error())
		}
	}()

	strr := storer.New(storage, consumer, logg)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	logg.Info("calendar storer is running...")

	if err := strr.Start(ctx, config.Kafka.Topic); err != nil {
		if err != context.Canceled {
			logg.Error("storer error: " + err.Error())
		}
	}

	logg.Info("calendar storer shutting down...")
}
