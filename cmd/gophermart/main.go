package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"loyalty-ledger/internal/config"
	"loyalty-ledger/internal/handler"
	"loyalty-ledger/internal/repository"
	"loyalty-ledger/internal/service"
	"loyalty-ledger/pkg/auth"
)

func main() {
	cfg := config.Load()

	repos, err := repository.InitRepositories(cfg.DatabaseURI, nil)
	if err != nil {
		log.Fatalf("repository initialization failed: %v", err)
	}
	if repos.PGX != nil {
		defer repos.PGX.Close()
	}

	userService := service.NewUserService(repos.User)
	JWTConfig := auth.DefaultJWTConfigWithSecret(cfg.JWTSecret)
	userHandler := handler.NewUserHandler(userService, JWTConfig)

	orderService := service.NewOrderService(repos.Order)
	orderHandler := handler.NewOrderHandler(orderService)

	handlers := handler.Handlers{
		User:  userHandler,
		Order: orderHandler,
	}
	mux := handler.NewRouter(handlers, cfg)

	srv := &http.Server{
		Addr:    cfg.RunAddress,
		Handler: mux,
	}

	errChan := make(chan error, 1)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Starting server at %s", cfg.RunAddress)
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	select {
	case sig := <-stop:
		log.Printf("Received signal: %v. Shutting down server...", sig)
	case err := <-errChan:
		log.Fatalf("server error: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server exited gracefully")
}
