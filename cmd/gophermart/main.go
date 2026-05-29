package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/paulwwyvern/gophermart/internal/config"
	"github.com/paulwwyvern/gophermart/internal/handler"
	mwauth "github.com/paulwwyvern/gophermart/internal/handler/middleware/auth"
	mwlogger "github.com/paulwwyvern/gophermart/internal/handler/middleware/logger"
	userRepository "github.com/paulwwyvern/gophermart/internal/repository/user"
	"github.com/paulwwyvern/gophermart/internal/service/user"
	"github.com/paulwwyvern/gophermart/pkg/jwtparse"
	"go.uber.org/zap"
	"log"
	"net/http"
	"time"
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

	userRepo := userRepository.NewStorage()

	userService := user.NewService(userRepo, tokenParser)

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
