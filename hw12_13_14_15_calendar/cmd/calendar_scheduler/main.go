package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/inenagl/hw-Go-Prof/hw12_13_14_15_calendar/internal/logger"
	"github.com/inenagl/hw-Go-Prof/hw12_13_14_15_calendar/internal/queue"
	"github.com/inenagl/hw-Go-Prof/hw12_13_14_15_calendar/internal/scheduler"
	memorystorage "github.com/inenagl/hw-Go-Prof/hw12_13_14_15_calendar/internal/storage/memory"
	sqlstorage "github.com/inenagl/hw-Go-Prof/hw12_13_14_15_calendar/internal/storage/sql"
	_ "github.com/jackc/pgx/stdlib"
)

var configFile string

func main() {
	flag.StringVar(&configFile, "config", "/etc/calendar/config_scheduler.yaml", "Path to configuration file")
	flag.Parse()

	if flag.Arg(0) == "version" {
		printVersion()
		return
	}

	config, err := NewConfig(configFile)
	if err != nil {
		log.Fatalln(err)
	}

	logg, err := logger.New(
		config.Logger.Preset,
		config.Logger.Level,
		config.Logger.Encoding,
		config.Logger.OutputPaths,
		config.Logger.ErrorOutputPaths,
	)
	if err != nil {
		log.Fatalln(err)
	}
	defer logg.Sync()

	var storage scheduler.Storage
	switch config.Storage.Type {
	case StorageInmemoryType:
		storage = memorystorage.New()
	case StorageSQLType:
		storage = sqlstorage.New(
			config.Storage.Host,
			config.Storage.Port,
			config.Storage.DBName,
			config.Storage.User,
			config.Storage.Password,
			config.Storage.SSLMode,
			config.Storage.Timeout,
		)
	default:
		logg.Error(fmt.Sprintf("unprocessable storage type: \"%s\"\n", config.Storage.Type))
		logg.Sync()
		os.Exit(1) //nolint: gocritic
	}

	producer := queue.NewProducer(
		config.Producer.Host,
		config.Producer.Port,
		config.Producer.User,
		config.Producer.Password,
		config.Producer.ExchangeName,
		config.Producer.ExchangeType,
		config.Producer.RoutingKey,
		config.Producer.QueueName,
		config.Producer.QosCount,
	)

	app := scheduler.New(config.Scheduler.WorkCycle, config.Scheduler.Expiration, *logg, storage, producer)

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP,
	)
	defer cancel()

	go func() {
		<-ctx.Done()

		if err := app.Stop(); err != nil {
			logg.Error("failed to stop Scheduler: " + err.Error())
		}
	}()

	if err = app.Start(ctx); err != nil {
		logg.Error("failed to start Scheduler: " + err.Error())
		cancel()
	}
}
