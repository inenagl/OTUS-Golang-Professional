package main

import (
	"fmt"
	"path"
	"strings"

	"github.com/spf13/viper"
)

const EnvVarPrefix = "GOCLNDR"

const (
	Direct  = "direct"
	FanOut  = "fanout"
	Topic   = "topic"
	XCustom = "x-custom"
)

type Config struct {
	Sender   SenderConf
	Logger   LoggerConf
	Consumer ConsumerConf
}

type SenderConf struct {
	Threads int
}

type LoggerConf struct {
	Preset           string
	Level            string
	Encoding         string
	OutputPaths      []string
	ErrorOutputPaths []string
}

type ConsumerConf struct {
	Host         string
	Port         int
	User         string
	Password     string
	ExchangeName string
	ExchangeType string
	RoutingKey   string
	QueueName    string
	QosCount     int
	ConsumerTag  string
}

func NewConfig(filePath string) (Config, error) {
	config := Config{}

	dir, name := path.Split(filePath)
	fParts := strings.SplitN(name, ".", 2)

	viper.SetConfigName(fParts[0])
	viper.SetConfigType(fParts[1])
	viper.AddConfigPath(dir)
	err := viper.ReadInConfig()
	if err != nil {
		return config, err
	}

	viper.SetEnvPrefix(EnvVarPrefix)
	viper.AutomaticEnv()

	config.Sender = processSenderConf()
	if err != nil {
		return config, err
	}

	config.Logger, err = processLoggerConf()
	if err != nil {
		return config, err
	}

	config.Consumer, err = processConsumerConf()
	if err != nil {
		return config, err
	}

	return config, nil
}

func inStrArray(needle string, arr []string) bool {
	for _, v := range arr {
		if needle == v {
			return true
		}
	}
	return false
}

func getAllowedStringVal(field string, allowed []string) (string, error) {
	value := viper.GetString(field)
	if !inStrArray(value, allowed) {
		return "", fmt.Errorf(`invalid %s value: "%s", allowed values are %v`, field, value, allowed)
	}
	return value, nil
}

func processSenderConf() SenderConf {
	viper.SetDefault("sender.threads", 1)
	return SenderConf{
		Threads: viper.GetInt("sender.threads"),
	}
}

func processLoggerConf() (LoggerConf, error) {
	viper.SetDefault("preset", "prod")
	viper.SetDefault("level", "info")
	viper.SetDefault("encoding", "json")
	viper.SetDefault("outputPaths", []string{"stderr"})
	viper.SetDefault("errorOutputPaths", []string{"stderr"})

	conf := LoggerConf{}

	val, err := getAllowedStringVal("logger.preset", []string{"dev", "prod"})
	if err != nil {
		return conf, err
	}
	conf.Preset = val

	val, err = getAllowedStringVal(
		"logger.level",
		[]string{"debug", "info", "warn", "error", "dpanic", "panic", "fatal"},
	)
	if err != nil {
		return conf, err
	}
	conf.Level = val

	val, err = getAllowedStringVal("logger.encoding", []string{"console", "json"})
	if err != nil {
		return conf, err
	}
	conf.Encoding = val

	conf.OutputPaths = viper.GetStringSlice("logger.outputPaths")
	conf.ErrorOutputPaths = viper.GetStringSlice("logger.errorOutputPaths")

	return conf, nil
}

func processConsumerConf() (ConsumerConf, error) {
	conf := ConsumerConf{}
	conf.Host = viper.GetString("AMQPHost")
	conf.Port = viper.GetInt("AMQPPort")
	conf.User = viper.GetString("AMQPUser")
	conf.Password = viper.GetString("AMQPPassword")
	conf.ExchangeName = viper.GetString("consumer.exchangeName")
	conf.RoutingKey = viper.GetString("consumer.routingKey")
	conf.QueueName = viper.GetString("consumer.queueName")
	conf.ConsumerTag = viper.GetString("consumer.consumerTag")
	conf.QosCount = viper.GetInt("consumer.qosCount")

	val, err := getAllowedStringVal("consumer.exchangeType", []string{Direct, FanOut, Topic, XCustom})
	if err != nil {
		return conf, err
	}
	conf.ExchangeType = val

	return conf, nil
}
