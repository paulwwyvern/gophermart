package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
	"go.uber.org/zap"
)

type DatabaseConfig struct {
	DatabaseURI     string `env:"DATABASE_URI" yaml:"database_uri"`
	MigrationSource string `env:"MIGRATION_SOURCE" yaml:"migration_source"`

	Host     string `env:"DB_HOST" yaml:"host"`
	Port     string `env:"DB_PORT" yaml:"port"`
	Username string `env:"DB_USERNAME" yaml:"username"`
	Password string `env:"DB_PASSWORD" yaml:"password"`
	Database string `env:"DB_DATABASE" yaml:"database"`
}
type Config struct {
	ConfigPath string

	RunAddress           string `env:"RUN_ADDRESS" yaml:"run_address"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS" yaml:"accrual_system_address"`

	DatabaseConfig DatabaseConfig `yaml:"database"`
}

func parseFlags() *Config {
	conf := &Config{}
	flag.StringVar(&conf.ConfigPath, "c", "", "Path to config file")

	flag.StringVar(&conf.RunAddress, "a", "", "Server address")
	flag.StringVar(&conf.DatabaseConfig.DatabaseURI, "d", "", "Database URI")
	flag.StringVar(&conf.AccrualSystemAddress, "r", "", "Accrual System Address")

	flag.Parse()

	return conf
}

func applyFlags(conf *Config, flagConf *Config) {
	if flagConf.RunAddress != "" {
		conf.RunAddress = flagConf.RunAddress
	}
	if flagConf.DatabaseConfig.DatabaseURI != "" {
		conf.DatabaseConfig.DatabaseURI = flagConf.DatabaseConfig.DatabaseURI
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

	conf.ConfigPath = confPath

	if _, err := os.Stat(confPath); os.IsNotExist(err) {
		// читаем из env если файлов нет
		cleanenv.ReadEnv(conf)
	} else {
		// читаем из файла
		cleanenv.ReadConfig(confPath, conf)
	}

	// наивысший приоритет у флагов
	applyFlags(conf, flagConf)

	// если database uri не задан:
	if conf.DatabaseConfig.DatabaseURI == "" {
		GetDatabaseURI(&conf.DatabaseConfig)
	}

	return conf
}

func GetDatabaseURI(dbc *DatabaseConfig) {
	if dbc.Host == "" || dbc.Database == "" || dbc.Username == "" || dbc.Password == "" || dbc.Port == "" {
		dbc.DatabaseURI = ""
		return
	}

	dbc.DatabaseURI = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbc.Username,
		dbc.Password,
		dbc.Host,
		dbc.Port,
		dbc.Database,
	)
	return
}

func LoggingConfig(logger *zap.Logger, conf *Config) {
	logger.Info(
		"Service configuration:",
		zap.String("Config path", conf.ConfigPath),
		zap.String("Run address", conf.RunAddress),
		zap.String("Database URI", conf.DatabaseConfig.DatabaseURI),
	)
}
