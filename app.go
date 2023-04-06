package main

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type collectorI interface {
	Update(ctx context.Context, ch chan<- prometheus.Metric) error
	Name() string
}

type application struct {
	config *config
}

func newApplication(config *config) *application {
	return &application{config: config}
}

func (app *application) start(ctx context.Context) error {
	if errRegister := prometheus.Register(newRootCollector(ctx, app.config)); errRegister != nil {
		log.Fatal("Couldn't register collector", zap.Error(errRegister))
	}
	handler := promhttp.HandlerFor(prometheus.DefaultGatherer,
		promhttp.HandlerOpts{ErrorHandling: promhttp.ContinueOnError})
	http.HandleFunc(app.config.MetricsPath, func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r)
	})
	return http.ListenAndServe(app.config.Addr(), nil)
}

func toFloat64Default(s string, def float64) float64 {
	n, errConv := strconv.ParseFloat(s, 64)
	if errConv != nil {
		log.Error("Failed to parse float64", zap.String("value", s))
		return def
	}
	return n
}

func toIntDefault(s string, def int) int {
	n, errConv := strconv.ParseInt(s, 10, 32)
	if errConv != nil {
		log.Error("Failed to parse int64", zap.String("value", s))
		return def
	}
	return int(n)
}
