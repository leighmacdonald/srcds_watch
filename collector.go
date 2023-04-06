package main

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"sync"
	"time"
)

type rootCollector struct {
	ctx             context.Context
	statsCollector  collectorI
	statusCollector collectorI
}

func newRootCollector(ctx context.Context, config *config) *rootCollector {
	return &rootCollector{
		ctx:             ctx,
		statsCollector:  newStatsCollector(config),
		statusCollector: newStatusCollector(config),
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
	collectors := []collectorI{n.statsCollector, n.statusCollector}
	wg := sync.WaitGroup{}
	wg.Add(len(collectors))
	for _, coll := range collectors {
		go func(coll collectorI) {
			defer wg.Done()
			c, cancel := context.WithTimeout(n.ctx, time.Second*10)
			defer cancel()
			if errUpdate := coll.Update(c, metricsCh); errUpdate != nil {
				log.Error("Failed to update collector", zap.Error(errUpdate))
			}

		}(coll)
	}
	wg.Wait()

	close(metricsCh)
	wgOut.Wait()
}
