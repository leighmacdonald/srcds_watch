package main

import (
	"context"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

type rootCollector struct {
	// ctx cant get passed via update call as it's not in the defined prom interface so its here until
	// background updates are enabled
	ctx             context.Context //nolint:containedctx
	statusCollector collectorI
	log             *zap.Logger
}

func newRootCollector(ctx context.Context, config *config, logger *zap.Logger, cm *connManager) *rootCollector {
	return &rootCollector{
		ctx:             ctx,
		statusCollector: newStatusCollector(config, cm, logger),
		log:             logger.Named("collector"),
	}
}

func (n *rootCollector) Name() string {
	return "srcds_watch"
}

func (n *rootCollector) Describe(_ chan<- *prometheus.Desc) {
}

func (n *rootCollector) Collect(outgoingCh chan<- prometheus.Metric) {
	metricsCh := make(chan prometheus.Metric)

	wgOut := sync.WaitGroup{}
	wgOut.Add(1)

	go func() {
		for metric := range metricsCh {
			outgoingCh <- metric
		}

		wgOut.Done()
	}()

	collectors := []collectorI{n.statusCollector}
	waitGroup := sync.WaitGroup{}

	waitGroup.Add(len(collectors))

	for _, coll := range collectors {
		go func(coll collectorI) {
			defer waitGroup.Done()

			c, cancel := context.WithTimeout(n.ctx, time.Second*10)
			defer cancel()

			if errUpdate := coll.Update(c, metricsCh); errUpdate != nil {
				n.log.Error("Failed to update collector", zap.Error(errUpdate))
			}
		}(coll)
	}

	waitGroup.Wait()
	close(metricsCh)
	wgOut.Wait()
}
