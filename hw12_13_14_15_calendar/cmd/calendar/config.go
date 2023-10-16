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

// При желании конфигурацию можно вынести в internal/config.
// Организация конфига в main принуждает нас сужать API компонентов, использовать
// при их конструировании только необходимые параметры, а также уменьшает вероятность циклической зависимости.
type Config struct {
	Logger     LoggerConf
	Storage    StorageConf
	HTTPServer ServerConf
	GRPCServer ServerConf
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

type ServerConf struct {
	Host string
	Port int
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

	config.Logger, err = processLoggerConf()
	if err != nil {
		return config, err
	}

	config.Storage, err = processStorageConf()
	if err != nil {
		return config, err
	}

	config.HTTPServer = processHTTPServerConf()
	config.GRPCServer = processGRPCServerConf()

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

	val, err = getAllowedStringVal("logger.level", []string{"debug", "info", "warn", "error", "dpanic", "panic", "fatal"})
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

func processHTTPServerConf() ServerConf {
	viper.SetDefault("host", "localhost")
	viper.SetDefault("port", 8080)

	conf := ServerConf{}
	conf.Host = viper.GetString("http-server.host")
	conf.Port = viper.GetInt("http-server.port")

	return conf
}

func processGRPCServerConf() ServerConf {
	viper.SetDefault("host", "localhost")
	viper.SetDefault("port", 8080)

	conf := ServerConf{}
	conf.Host = viper.GetString("grpc-server.host")
	conf.Port = viper.GetInt("grpc-server.port")

	return conf
}
