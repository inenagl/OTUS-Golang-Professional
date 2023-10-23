package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"

	"github.com/inenagl/hw-Go-Prof/hw12_13_14_15_calendar/internal/logger"
	"github.com/inenagl/hw-Go-Prof/hw12_13_14_15_calendar/internal/queue"
	"github.com/inenagl/hw-Go-Prof/hw12_13_14_15_calendar/internal/sender"
	_ "github.com/jackc/pgx/stdlib"
)

var configFile string

func main() {
	flag.StringVar(&configFile, "config", "/etc/calendar/config_sender.yaml", "Path to configuration file")
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

	consumer := queue.NewConsumer(
		config.Consumer.Host,
		config.Consumer.Port,
		config.Consumer.User,
		config.Consumer.Password,
		config.Consumer.ExchangeName,
		config.Consumer.ExchangeType,
		config.Consumer.RoutingKey,
		config.Consumer.QueueName,
		config.Consumer.ConsumerTag,
		config.Consumer.QosCount,
	)

	app := sender.New(config.Sender.Threads, logg, consumer)

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP,
	)
	defer cancel()

	go func() {
		<-ctx.Done()

		if err := app.Stop(); err != nil {
			logg.Error("failed to stop Sender: " + err.Error())
		}
	}()

	if err = app.Start(ctx); err != nil {
		logg.Error("failed to start Sender: " + err.Error())
		cancel()
	}
}
