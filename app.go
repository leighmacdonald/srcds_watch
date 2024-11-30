package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	errPromRegister = errors.New("failed to register prometheus collection")
	errHTTPListen   = errors.New("HTTP listener returned error")
)

func start(ctx context.Context, config *config) error {
	if errRegister := prometheus.Register(newRootCollector(ctx, config)); errRegister != nil {
		return errors.Join(errRegister, errPromRegister)
	}

	handler := promhttp.HandlerFor(prometheus.DefaultGatherer,
		promhttp.HandlerOpts{ErrorHandling: promhttp.ContinueOnError})

	http.HandleFunc(config.MetricsPath, func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r)
	})

	httpServer := &http.Server{
		Addr:           config.Addr(),
		Handler:        nil,
		ReadTimeout:    20 * time.Second,
		WriteTimeout:   20 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		<-ctx.Done()

		slog.Info("Shutdown signal received")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		if errShutdown := httpServer.Shutdown(shutdownCtx); errShutdown != nil { //nolint:contextcheck
			slog.Error("Error shutting down http service", slog.String("error", errShutdown.Error()))
		}
	}()

	if errServe := httpServer.ListenAndServe(); errServe != nil && !errors.Is(errServe, http.ErrServerClosed) {
		return errors.Join(errServe, errHTTPListen)
	}

	slog.Info("Shutdown successful. Bye.")

	return nil
}

func toFloat64Default(s string, def float64) float64 {
	parsedValue, errConv := strconv.ParseFloat(s, 64)
	if errConv != nil {
		return def
	}

	return parsedValue
}

func toIntDefault(s string, def int) int {
	parsedValue, errConv := strconv.ParseInt(s, 10, 32)
	if errConv != nil {
		return def
	}

	return int(parsedValue)
}
