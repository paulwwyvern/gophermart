package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/paulwwyvern/gophermart/internal/config"
	"github.com/paulwwyvern/gophermart/internal/handler"
	mwauth "github.com/paulwwyvern/gophermart/internal/handler/middleware/auth"
	mwcompress "github.com/paulwwyvern/gophermart/internal/handler/middleware/compress"
	mwlogger "github.com/paulwwyvern/gophermart/internal/handler/middleware/logger"
	"github.com/paulwwyvern/gophermart/internal/repository/postgres"
	"github.com/paulwwyvern/gophermart/internal/service"
	"github.com/paulwwyvern/gophermart/pkg/jwtparse"
	"go.uber.org/zap"
)

const (
	tokenSecret = "abcde"
	tokenTTL    = time.Hour * 168
	maxBodyLen  = 1024 * 1024
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
	logger.Info("Migrate database successfully")

	storage, err := postgres.NewStorage(conf.DatabaseConfig.DatabaseURI)
	if err != nil {
		log.Fatalf("can't initialize postgres storage: %v", err)
	}
	defer storage.Close()
	logger.Info("Create postgres connection")

	userService := service.NewUserService(storage, tokenParser)
	logger.Info("Create user service successfully")

	orderService := service.NewOrderService(storage)
	logger.Info("Create order service successfully")

	balanceService := service.NewBalanceService(storage)
	logger.Info("Create balance service successfully")

	h := handler.NewHandler(maxBodyLen, userService, orderService, balanceService)

	r := chi.NewRouter()
	r.Use(mwlogger.WithLogger(logger))
	r.Use(mwcompress.WithCompress())

	r.Post("/api/user/register", h.RegisterUser)
	r.Post("/api/user/login", h.LoginUser)
	r.Group(func(r chi.Router) {
		r.Use(mwauth.WithAuth(tokenParser))

		r.Post("/api/user/orders", h.CreateOrder)
		r.Get("/api/user/orders", h.GetOrders)
		r.Post("/api/user/balance", h.AddBalance)
		r.Get("/api/user/balance", h.GetBalance)
		r.Post("/api/user/balance/withdraw", h.CreateWithdrawal)
		r.Get("/api/user/withdrawals", h.GetWithdrawals)
		r.Get("/echo", h.Echo)
	})
	logger.Info("Initialize handler successfully")

	server := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
