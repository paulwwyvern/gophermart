package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/paulwwyvern/gophermart/internal/config"
	"github.com/paulwwyvern/gophermart/internal/handler"
	mwauth "github.com/paulwwyvern/gophermart/internal/handler/middleware/auth"
	mwcompress "github.com/paulwwyvern/gophermart/internal/handler/middleware/compress"
	mwlogger "github.com/paulwwyvern/gophermart/internal/handler/middleware/logger"
	"github.com/paulwwyvern/gophermart/internal/repository/accrual"
	"github.com/paulwwyvern/gophermart/internal/repository/postgres"
	"github.com/paulwwyvern/gophermart/internal/service"
	"github.com/paulwwyvern/gophermart/pkg/jwtparse"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

const (
	maxBodyLen = 1024 * 1024
)

func main() {
	decimal.MarshalJSONWithoutQuotes = true

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	conf := config.NewConfig()

	// Конфигурация логгера
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	defer logger.Sync()

	// Вывод конфигурации
	config.LoggingConfig(logger, conf)

	tokenParser := jwtparse.NewParser(conf.TokenConfig.Secret, conf.TokenConfig.TTL)

	// инициальзация postgres
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

	accrualClient := accrual.NewClient(conf.AccrualSystemAddress)

	userService := service.NewUserService(storage, tokenParser)
	logger.Info("Create user service successfully")

	orderService := service.NewOrderService(storage)
	logger.Info("Create order service successfully")

	balanceService := service.NewBalanceService(storage)
	logger.Info("Create balance service successfully")

	orderWorkerPool := service.NewOrderWorkerPool(accrualClient, storage,
		conf.WorkersConfig.Count, conf.WorkersConfig.BatchSize, conf.WorkersConfig.RetryAfterDefault)
	orderWorkerPool.Run(ctx, logger, conf.WorkersConfig.PollInterval)
	defer orderWorkerPool.Wait()

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
	})
	logger.Info("Initialize handler successfully")

	server := &http.Server{
		Addr:    ":8080",
		Handler: r,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()
	logger.Info("Server successfully started")

	<-ctx.Done()
	logger.Info("Shutting down server...")
	stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	err = server.Shutdown(shutdownCtx)
	cancel()
	if err != nil {
		logger.Info("Shutdown server err", zap.Error(err))
	}
	logger.Info("Server successfully shutdown")
}
