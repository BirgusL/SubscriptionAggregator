// @title Subscription Aggregator API
// @version 1.0
// @description API для управления подписками пользователей

// @contact.name Kuzmin Anton
// @contact.email kuzmin1a.a@gmail.com

// @host localhost:8080
// @BasePath /

package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"

	_ "SubscriptionAggregator/docs"
	"SubscriptionAggregator/pkg/config"
	"SubscriptionAggregator/pkg/handler"
	"SubscriptionAggregator/pkg/repository"
	"SubscriptionAggregator/pkg/service"

	httpSwagger "github.com/swaggo/http-swagger"
)

const (
	envLocal  = "local"
	envDocker = "docker"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info("starting subscriptionaggregator", slog.String("env", cfg.Env))
	log.Debug("debug messages are enabled")

	dbURL := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.Name,
		cfg.Sslmode,
	)

	pg, err := repository.New(ctx, dbURL)
	if err != nil {
		log.Error("failed to initialize database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pg.Close()

	repo := repository.NewSubscriptionRepository(pg.DB)

	router := mux.NewRouter()
	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	svc := service.NewSubscriptionService(repo)

	hlr := handler.NewSubscriptionHandler(svc)

	hlr.RegisterRoutes(router)

	srv := &http.Server{
		Addr:         cfg.Adress,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.TimeOut,
		WriteTimeout: cfg.HTTPServer.TimeOut,
		IdleTimeout:  cfg.HTTPServer.IdleTimeOut,
	}

	done := make(chan os.Signal, 1)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("failed to start server", slog.String("error", err.Error()))
		}
	}()
	log.Info("server started", slog.String("adress", cfg.Adress))

	<-done
	log.Info("server stopped")

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("server shutdown failed", slog.String("error", err.Error()))
	}
	log.Info("server exited properly")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envDocker:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}

	return log
}
