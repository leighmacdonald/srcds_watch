package main

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"sync"
	"time"
)

type RootCollector struct {
	ctx            context.Context
	statsCollector collectorI
}

func newRootCollector(ctx context.Context, config *config) *RootCollector {
	return &RootCollector{
		ctx:            ctx,
		statsCollector: newStatsCollector(config),
	}
}
func (n *RootCollector) Name() string {
	return "srcds_watch"
}

func (n *RootCollector) Describe(_ chan<- *prometheus.Desc) {
}

func (n *RootCollector) Collect(outgoingCh chan<- prometheus.Metric) {
	metricsCh := make(chan prometheus.Metric)

	wgOutgoing := sync.WaitGroup{}
	wgOutgoing.Add(1)
	go func() {
		for metric := range metricsCh {
			outgoingCh <- metric
		}
		wgOutgoing.Done()
	}()
	collectors := []collectorI{n.statsCollector}
	wgCollection := sync.WaitGroup{}
	wgCollection.Add(len(collectors))
	for _, coll := range collectors {
		go func(coll collectorI) {
			c, cancel := context.WithTimeout(n.ctx, time.Second*10)
			defer cancel()
			if errUpdate := coll.Update(metricsCh, c); errUpdate != nil {
				log.Error("Failed to update collector", zap.Error(errUpdate))
				return
			}
			wgCollection.Done()
		}(coll)
	}
	wgCollection.Wait()

	close(metricsCh)
	wgOutgoing.Wait()
}
