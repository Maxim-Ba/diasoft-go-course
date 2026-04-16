package main

import (
	"context"
	"flag"
	"os/signal"
	"syscall"
	"time"

	"github.com/Maxim-Ba/hw12_13_14_15_calendar/internal/logger"
	"github.com/Maxim-Ba/hw12_13_14_15_calendar/internal/queue"
	"github.com/Maxim-Ba/hw12_13_14_15_calendar/internal/scheduler"
	sqlstorage "github.com/Maxim-Ba/hw12_13_14_15_calendar/internal/storage/sql"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "configs/scheduler_config.yaml", "Path to configuration file")
}

func main() {
	flag.Parse()

	config, err := NewConfig(configFile)
	if err != nil {
		logger.New("ERROR").Error("failed to load config: " + err.Error())
		panic(err.Error())
	}

	logg := logger.New(config.Logger.Level)
	logg.Info("calendar scheduler starting...")

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

	producer, err := queue.NewKafkaProducer(config.Kafka.Brokers)
	if err != nil {
		logg.Error("failed to create Kafka producer: " + err.Error())
		if closeErr := storage.Close(); closeErr != nil {
			logg.Error("failed to close SQL storage: " + closeErr.Error())
		}
		panic(err.Error())
	}
	defer func() {
		if err := producer.Close(); err != nil {
			logg.Error("failed to close Kafka producer: " + err.Error())
		}
	}()

	sched := scheduler.New(storage, producer, config.Kafka.Topic, logg)

	scanInterval, err := time.ParseDuration(config.Schedule.ScanInterval)
	if err != nil {
		logg.Error("invalid scan interval: " + err.Error())
		panic(err.Error())
	}

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	ticker := time.NewTicker(scanInterval)
	defer ticker.Stop()

	cleanupTicker := time.NewTicker(24 * time.Hour)
	defer cleanupTicker.Stop()

	logg.Info("calendar scheduler is running...")

	for {
		select {
		case <-ctx.Done():
			logg.Info("calendar scheduler shutting down...")
			return
		case <-ticker.C:
			logg.Info("scanning for events to notify...")
			if err := sched.ScanAndNotify(ctx); err != nil {
				logg.Error("failed to scan and notify: " + err.Error())
			}
		case <-cleanupTicker.C:
			logg.Info("cleaning up old events...")
			if err := sched.CleanupOldEvents(ctx); err != nil {
				logg.Error("failed to cleanup old events: " + err.Error())
			}
		}
	}
}
