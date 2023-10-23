package main

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/spf13/viper"
)

const EnvVarPrefix = "GOCLNDR"

const (
	StorageInmemoryType = "memory"
	StorageSQLType      = "sql"
)

const (
	Direct  = "direct"
	FanOut  = "fanout"
	Topic   = "topic"
	XCustom = "x-custom"
)

type Config struct {
	Scheduler SchedulerConf
	Logger    LoggerConf
	Storage   StorageConf
	Producer  ProducerConf
}

type SchedulerConf struct {
	WorkCycle  time.Duration
	Expiration time.Duration
}

type LoggerConf struct {
	Preset           string
	Level            string
	Encoding         string
	OutputPaths      []string
	ErrorOutputPaths []string
}

type StorageConf struct {
	Type     string
	Host     string
	Port     int
	DBName   string
	User     string
	Password string
	SSLMode  string
	Timeout  time.Duration
}

type ProducerConf struct {
	Host         string
	Port         int
	User         string
	Password     string
	ExchangeName string
	ExchangeType string
	RoutingKey   string
	QueueName    string
	QosCount     int
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

	config.Scheduler = processSchedulerConf()

	config.Logger, err = processLoggerConf()
	if err != nil {
		return config, err
	}

	config.Storage, err = processStorageConf()
	if err != nil {
		return config, err
	}

	config.Producer, err = processProducerConf()
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

func processSchedulerConf() SchedulerConf {
	viper.SetDefault("scheduler.workcycle", time.Minute)
	viper.SetDefault("scheduler.expiration", 24*365*time.Hour)
	return SchedulerConf{
		WorkCycle:  viper.GetDuration("scheduler.workcycle"),
		Expiration: viper.GetDuration("scheduler.expiration"),
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

func processStorageConf() (StorageConf, error) {
	viper.SetDefault("type", StorageInmemoryType)
	viper.SetDefault("DBSSLMode", "require")
	viper.SetDefault("DBTimeout", time.Second*3)

	conf := StorageConf{}

	val, err := getAllowedStringVal("storage.type", []string{StorageInmemoryType, StorageSQLType})
	if err != nil {
		return conf, err
	}
	conf.Type = val

	if conf.Type == StorageSQLType {
		if !viper.IsSet("DBHost") || !viper.IsSet("DBName") || !viper.IsSet("DbUser") || !viper.IsSet("DbPassword") {
			return conf, fmt.Errorf("not all database requisites are set, need set dbhost, dbname, dbuser, dbpassword")
		}
		conf.Host = viper.GetString("DBHost")
		conf.Port = viper.GetInt("DBPort")
		conf.DBName = viper.GetString("DBName")
		conf.User = viper.GetString("DBUser")
		conf.Password = viper.GetString("DBPassword")
		conf.SSLMode = viper.GetString("DBSSLMode")
		conf.Timeout = viper.GetDuration("DBTimeout")
	}

	return conf, nil
}

func processProducerConf() (ProducerConf, error) {
	conf := ProducerConf{}
	conf.Host = viper.GetString("AMQPHost")
	conf.Port = viper.GetInt("AMQPPort")
	conf.User = viper.GetString("AMQPUser")
	conf.Password = viper.GetString("AMQPPassword")
	conf.ExchangeName = viper.GetString("producer.exchangeName")
	conf.RoutingKey = viper.GetString("producer.routingKey")
	conf.QueueName = viper.GetString("producer.queueName")
	conf.QosCount = viper.GetInt("producer.qosCount")

	val, err := getAllowedStringVal("producer.exchangeType", []string{Direct, FanOut, Topic, XCustom})
	if err != nil {
		return conf, err
	}
	conf.ExchangeType = val

	return conf, nil
}
