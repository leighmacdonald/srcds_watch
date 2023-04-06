package main

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"sync"
	"time"
)

type RootCollector struct {
	ctx             context.Context
	statsCollector  collectorI
	statusCollector collectorI
}

func newRootCollector(ctx context.Context, config *config) *RootCollector {
	return &RootCollector{
		ctx:             ctx,
		statsCollector:  newStatsCollector(config),
		statusCollector: newStatusCollector(config),
	}
}
func (n *RootCollector) Name() string {
	return "srcds_watch"
}

func (n *RootCollector) Describe(_ chan<- *prometheus.Desc) {
}

func (n *RootCollector) Collect(outgoingCh chan<- prometheus.Metric) {
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
			if errUpdate := coll.Update(metricsCh, c); errUpdate != nil {
				log.Error("Failed to update collector", zap.Error(errUpdate))
			}

		}(coll)
	}
	wg.Wait()

	close(metricsCh)
	wgOut.Wait()
}
