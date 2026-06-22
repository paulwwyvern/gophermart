package config

import (
	"crypto/rand"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type DatabaseConfig struct {
	DatabaseURI     string `env:"DATABASE_URI" yaml:"database_uri"`
	MigrationSource string `env:"MIGRATION_SOURCE" yaml:"migration_source" env-default:"./migration"`

	Host     string `env:"DB_HOST" yaml:"host"`
	Port     string `env:"DB_PORT" yaml:"port"`
	Username string `env:"DB_USERNAME" yaml:"username"`
	Password string `env:"DB_PASSWORD" yaml:"password"`
	Database string `env:"DB_DATABASE" yaml:"database"`
}

type TokenConfig struct {
	Secret string        `env:"TOKEN_SECRET" yaml:"secret"`
	TTL    time.Duration `env:"TOKEN_TTL" yaml:"ttl" env-default:"168h"`
}

type WorkersConfig struct {
	Count             int           `env:"WORKERS_COUNT" yaml:"count" env-default:"1"`
	BatchSize         int           `env:"WORKERS_BATCH_SIZE" yaml:"batch_size" env-default:"10"`
	RetryAfterDefault int           `env:"WORKERS_RETRY_AFTER_DEFAULT" yaml:"retry_after_default" env-default:"10"`
	PollInterval      time.Duration `env:"WORKERS_POLL_INTERVAL" yaml:"poll_interval" env-default:"10s"`
}

type Config struct {
	ConfigPath string

	RunAddress           string `env:"RUN_ADDRESS" yaml:"run_address"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS" yaml:"accrual_system_address"`

	DatabaseConfig DatabaseConfig `yaml:"database"`
	TokenConfig    TokenConfig    `yaml:"token"`
	WorkersConfig  WorkersConfig  `yaml:"workers"`
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

func NewConfig() (*Config, error) {
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

	// Чтение конфига из файлов
	err := cleanenv.ReadConfig(confPath, conf)
	if err != nil {
		// если чтение из конфигов из файла завершилось ошибкой
		err = cleanenv.ReadEnv(conf)
	}

	if err != nil {
		return nil, err
	}

	// наивысший приоритет у флагов
	applyFlags(conf, flagConf)

	// если database uri не задан:
	if conf.DatabaseConfig.DatabaseURI == "" {
		GetDatabaseURI(&conf.DatabaseConfig)
	}

	// если токен не задан
	if conf.TokenConfig.Secret == "" {
		conf.TokenConfig.Secret = rand.Text()
	}

	return conf, nil
}

func GetDatabaseURI(dbc *DatabaseConfig) {
	if dbc.Host == "" || dbc.Database == "" || dbc.Username == "" || dbc.Port == "" {
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

func LoggingConfig(logger *slog.Logger, conf *Config) {
	logger.Info(
		"Service configuration:",
		slog.String("Config path", conf.ConfigPath),
		slog.String("Run address", conf.RunAddress),
		slog.String("Database URI", conf.DatabaseConfig.DatabaseURI),
		slog.String("Accrual System Address", conf.AccrualSystemAddress),
	)
}
