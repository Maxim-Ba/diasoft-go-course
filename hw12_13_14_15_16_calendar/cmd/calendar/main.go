package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Maxim-Ba/hw12_13_14_15_calendar/internal/app"
	"github.com/Maxim-Ba/hw12_13_14_15_calendar/internal/logger"
	internalhttp "github.com/Maxim-Ba/hw12_13_14_15_calendar/internal/server/http"
	memorystorage "github.com/Maxim-Ba/hw12_13_14_15_calendar/internal/storage/memory"
	sqlstorage "github.com/Maxim-Ba/hw12_13_14_15_calendar/internal/storage/sql"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "configs/config.yaml", "Path to configuration file")
}

func main() {
	flag.Parse()

	if flag.Arg(0) == "version" {
		printVersion()
		return
	}

	config, err := NewConfig(configFile)
	if err != nil {
		logger.New("ERROR").Error("failed to load config: " + err.Error())
		os.Exit(1)
	}

	logg := logger.New(config.Logger.Level)

	var storage interface {
		app.Storage
	}

	switch config.Storage.Type {
	case "memory":
		logg.Info("using in-memory storage")
		storage = memorystorage.New()
	case "sql":
		logg.Info("using SQL storage")
		logg.Info("connecting to database: " + config.Database.DSN)

		sqlStorage, err := sqlstorage.New(config.Database.DSN)
		if err != nil {
			logg.Error("failed to create SQL storage: " + err.Error())
			os.Exit(1)
		}
		defer sqlStorage.Close()

		logg.Info("database migrations applied successfully")
		storage = sqlStorage
	default:
		logg.Error("unknown storage type: " + config.Storage.Type)
		os.Exit(1)
	}

	calendar := app.New(logg, storage)

	server := internalhttp.NewServer(logg, calendar)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		if err := server.Stop(ctx); err != nil {
			logg.Error("failed to stop http server: " + err.Error())
		}
	}()

	logg.Info("calendar is running...")

	if err := server.Start(ctx); err != nil {
		logg.Error("failed to start http server: " + err.Error())
		cancel()
		os.Exit(1) //nolint:gocritic
	}
}
