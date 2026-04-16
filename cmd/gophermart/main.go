package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"loyalty-ledger/internal/accrual"
	"loyalty-ledger/internal/config"
	"loyalty-ledger/internal/handler"
	"loyalty-ledger/internal/repository"
	"loyalty-ledger/internal/service"
	"loyalty-ledger/pkg/auth"
)

func main() {
	if err := run(); err != nil {
		log.Printf("fatal: %v", err)
		os.Exit(1)
	}
}

func run() error {
	cfg := config.Load()

	repos, err := repository.InitRepositories(cfg.DatabaseURI, nil)
	if err != nil {
		return fmt.Errorf("repository initialization failed: %w", err)
	}
	if repos.PGX != nil {
		defer repos.PGX.Close()
	}

	userService := service.NewUserService(repos.User)
	jwtConfig := auth.DefaultJWTConfigWithSecret(cfg.JWTSecret)
	userHandler := handler.NewUserHandler(userService, jwtConfig)

	orderService := service.NewOrderService(repos.Order)
	orderHandler := handler.NewOrderHandler(orderService)

	balanceService := service.NewBalanceService(repos.Balance)
	balanceHandler := handler.NewBalanceHandler(balanceService)

	accrualClient := accrual.NewHTTPClient(cfg.AccrualAddr)
	accrualInterval := time.Duration(cfg.AccrualPoller) * time.Second
	accrualService := service.NewAccrualService(repos.Order, repos.Balance)

	handlers := handler.Handlers{
		User:    userHandler,
		Order:   orderHandler,
		Balance: balanceHandler,
	}
	mux := handler.NewRouter(handlers, cfg)

	srv := &http.Server{
		Addr:         cfg.RunAddress,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	log.Println("Starting accrual workers and server...")
	errChan := make(chan error, 1)
	wg.Go(func() {
		log.Printf("Starting server at %s", cfg.RunAddress)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	})
	wg.Go(func() {
		accrual.StartNewToProcessingWorker(ctx, accrualService, accrualInterval)
	})
	wg.Go(func() {
		accrual.StartProcessingAccrualWorker(ctx, accrualClient, accrualService, accrualInterval, cfg.AccrualWorkers)
	})

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	var serverErr error
	select {
	case sig := <-stop:
		log.Printf("Received shutdown signal: %v. Shutting down...", sig)
	case serverErr = <-errChan:
		log.Printf("Server error: %v. Shutting down...", serverErr)
	}
	signal.Stop(stop)

	// Завершаем работу воркеров (перестают брать новые задачи)
	cancel()

	// Завершаем работу сервера (5 секунд на обработку активных запросов)
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	// Ожидаем завершения всех горутин
	wg.Wait()
	log.Println("Server and workers exited gracefully")

	return serverErr
}
