package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/paulwwyvern/gophermart/internal/config"
	"github.com/paulwwyvern/gophermart/internal/handler"
	mwauth "github.com/paulwwyvern/gophermart/internal/handler/middleware/auth"
	mwlogger "github.com/paulwwyvern/gophermart/internal/handler/middleware/logger"
	"github.com/paulwwyvern/gophermart/internal/repository/postgres"
	"github.com/paulwwyvern/gophermart/internal/service"
	"github.com/paulwwyvern/gophermart/pkg/jwtparse"
	"go.uber.org/zap"
)

const (
	tokenSecret = "abcde"
	tokenTTL    = time.Second * 10
)

func main() {
	conf := config.NewConfig()

	// Конфигурация логгера
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	defer logger.Sync()

	// Вывод конфигурации
	config.LoggingConfig(logger, conf)

	tokenParser := jwtparse.NewParser(tokenSecret, tokenTTL)

	err = postgres.Migrate(conf.DatabaseConfig.MigrationSource, conf.DatabaseConfig.DatabaseURI)
	if err != nil {
		log.Fatalf("can't migrate database: %v", err)
	}

	storage, err := postgres.NewStorage(conf.DatabaseConfig.DatabaseURI)
	if err != nil {
		log.Fatalf("can't initialize postgres storage: %v", err)
	}
	defer storage.Close()

	userService := service.NewUserService(storage, tokenParser)

	h := handler.NewHandler(userService)

	r := chi.NewRouter()
	r.Use(mwlogger.WithLogger(logger))
	r.Post("/api/user/register", h.RegisterUser)
	r.Post("/api/user/login", h.LoginUser)
	r.Group(func(r chi.Router) {
		r.Use(mwauth.WithAuth(tokenParser))

		r.Get("/echo", h.Echo)
	})

	server := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
