package main

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

type collectorI interface {
	Update(ctx context.Context, ch chan<- prometheus.Metric) error
	Name() string
}

type application struct {
	config *config
	log    *zap.Logger
	cm     *connManager
}

func newApplication(config *config, logger *zap.Logger) *application {
	return &application{config: config, log: logger.Named("srcds_watch"), cm: newConnManager()}
}

func (app *application) start(ctx context.Context) error {
	if errRegister := prometheus.Register(newRootCollector(ctx, app.config, app.log, app.cm)); errRegister != nil {
		app.log.Fatal("Couldn't register collector", zap.Error(errRegister))
	}

	handler := promhttp.HandlerFor(prometheus.DefaultGatherer,
		promhttp.HandlerOpts{ErrorHandling: promhttp.ContinueOnError})

	http.HandleFunc(app.config.MetricsPath, func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r)
	})

	httpServer := &http.Server{
		Addr:           app.config.Addr(),
		Handler:        nil,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   120 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		<-ctx.Done()

		app.log.Info("Shutdown signal received")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		if errShutdown := httpServer.Shutdown(shutdownCtx); errShutdown != nil { //nolint:contextcheck
			app.log.Error("Error shutting down http service", zap.Error(errShutdown))
		}
	}()

	if errServe := httpServer.ListenAndServe(); errServe != nil && !errors.Is(errServe, http.ErrServerClosed) {
		return errors.Wrap(errServe, "HTTP listener returned error")
	}

	app.log.Info("Shutdown successful. Bye.")

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
