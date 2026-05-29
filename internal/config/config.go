package config

import (
	"flag"
	"github.com/ilyakaznacheev/cleanenv"
	"go.uber.org/zap"
	"os"
)

type Config struct {
	ConfigPath string

	RunAddress           string `env:"RUN_ADDRESS" yaml:"run_address"`
	DatabaseURI          string `env:"DATABASE_URI" yaml:"database_uri"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS" yaml:"accrual_system_address"`
}

func parseFlags() *Config {
	conf := &Config{}
	flag.StringVar(&conf.ConfigPath, "c", "", "Path to config file")

	flag.StringVar(&conf.RunAddress, "a", "", "Server address")
	flag.StringVar(&conf.DatabaseURI, "d", "", "Database URI")
	flag.StringVar(&conf.AccrualSystemAddress, "r", "", "Accrual System Address")

	flag.Parse()

	return conf
}

func applyFlags(conf *Config, flagConf *Config) {
	if flagConf.RunAddress != "" {
		conf.RunAddress = flagConf.RunAddress
	}
	if flagConf.DatabaseURI != "" {
		conf.DatabaseURI = flagConf.DatabaseURI
	}
	if flagConf.AccrualSystemAddress != "" {
		conf.AccrualSystemAddress = flagConf.AccrualSystemAddress
	}
}

func NewConfig() *Config {
	conf := &Config{}
	confPath := ""

	// Получаем конфиги из флагов
	flagConf := parseFlags()

	// Получаем путь к файлу из env
	envConfPath, ok := os.LookupEnv("CONFIG_PATH")
	if ok {
		confPath = envConfPath
	}

	// Получаем путь к файлу из флагов
	if flagConf.ConfigPath != "" {
		confPath = flagConf.ConfigPath
	}

	if _, err := os.Stat(confPath); os.IsNotExist(err) {
		// читаем из env если файлов нет
		cleanenv.ReadEnv(conf)
	} else {
		// читаем из файла
		cleanenv.ReadConfig(confPath, conf)
	}

	// наивысший приоритет у флагов
	applyFlags(conf, flagConf)

	return conf
}

func LoggingConfig(logger *zap.Logger, conf *Config) {
	logger.Info(
		"Service configuration:",
		zap.String("Config path", conf.ConfigPath),
		zap.String("Run address", conf.RunAddress),
		zap.String("Database URI", conf.DatabaseURI),
	)
}
