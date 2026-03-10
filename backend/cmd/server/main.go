// Package main is the entry point for the phone-call-receptionist backend server.
package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"

	"phone-call-receptionist/backend/internal/bootstrap"
	"phone-call-receptionist/backend/internal/configuration"
)

// @title           Phone Call Receptionist API
// @version         1.0
// @description     Voice AI Receptionist for IT services firm
// @host            localhost:8080
// @BasePath        /api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Caller().Logger()
	if configuration.Config.Server.Env == "development" {
		logger = logger.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	}

	logger.Info().Str("port", configuration.Config.Server.Port).Msg("Starting server")

	app, err := bootstrap.Initialize(&logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize application")
	}
	defer app.Close()

	srv := &http.Server{
		Addr:         ":" + configuration.Config.Server.Port,
		Handler:      app.Router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("Server failed")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info().Msg("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	logger.Info().Msg("Server exited gracefully")
}
