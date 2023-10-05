package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/inenagl/hw-Go-Prof/hw12_13_14_15_calendar/internal/app"
	"github.com/inenagl/hw-Go-Prof/hw12_13_14_15_calendar/internal/logger"
	internalhttp "github.com/inenagl/hw-Go-Prof/hw12_13_14_15_calendar/internal/server/http"
	memorystorage "github.com/inenagl/hw-Go-Prof/hw12_13_14_15_calendar/internal/storage/memory"
	sqlstorage "github.com/inenagl/hw-Go-Prof/hw12_13_14_15_calendar/internal/storage/sql"
	_ "github.com/jackc/pgx/stdlib"
)

var configFile string

func main() {
	flag.StringVar(&configFile, "config", "/etc/calendar/config.yaml", "Path to configuration file")
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
		config.Logger.ErrorOutputPaths)
	if err != nil {
		log.Fatalln(err)
	}
	defer logg.Sync()

	var storage app.Storage
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
			config.Storage.Timeout)
	default:
		logg.Error(fmt.Sprintf("unprocessable storage type: \"%s\"\n", config.Storage.Type))
		logg.Sync()
		os.Exit(1) //nolint: gocritic
	}

	calendar := app.New(*logg, storage)

	server := internalhttp.NewServer(config.Server.Host, config.Server.Port, *logg, calendar)

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
		logg.Sync()
		os.Exit(1)
	}
}
