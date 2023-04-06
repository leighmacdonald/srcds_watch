package main

import (
	"context"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"regexp"
	"strings"
)

var (
	reStats *regexp.Regexp
)

type stats struct {
	CPU      float64
	NetIn    float64
	NetOut   float64
	Uptime   int
	Maps     int
	FPS      float64
	Player   int
	Connects int
}

type statsCollector struct {
	config   *config
	cpu      []*prometheus.Desc
	netIn    []*prometheus.Desc
	netOut   []*prometheus.Desc
	uptime   []*prometheus.Desc
	maps     []*prometheus.Desc
	fps      []*prometheus.Desc
	players  []*prometheus.Desc
	connects []*prometheus.Desc
}

func createStatsDesc(stat string, serverName string) *prometheus.Desc {
	switch stat {
	case "cpu":
		return prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "stats", stat),
			"The current cpu usage.",
			nil, prometheus.Labels{
				"server": serverName,
			})
	case "net_in":
		return prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "stats", stat),
			"The current inbound network traffic rate (KB/s)",
			nil, prometheus.Labels{
				"server": serverName,
			})
	case "net_out":
		return prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "stats", stat),
			"The current outbound network traffic rate (KB/s)",
			nil, prometheus.Labels{
				"server": serverName,
			})
	case "uptime":
		return prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "stats", stat),
			"The current server uptime in minutes",
			nil, prometheus.Labels{
				"server": serverName,
			})
	case "maps":
		return prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "stats", stat),
			"The total number of maps that have been played",
			nil, prometheus.Labels{
				"server": serverName,
			})
	case "fps":
		return prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "stats", stat),
			"The current server fps (tickrate)",
			nil, prometheus.Labels{
				"server": serverName,
			})
	case "players":
		return prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "stats", stat),
			"The current player count of the server.",
			nil, prometheus.Labels{
				"server": serverName,
			})
	case "connects":
		return prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "stats", stat),
			"The total number of players that have connected to the server.",
			nil, prometheus.Labels{
				"server": serverName,
			})
	default:
		log.Panic("Unhandled stat Name", zap.String("stat", stat))
	}
	return nil
}

func newStatsCollector(config *config) *statsCollector {
	var (
		cpu      []*prometheus.Desc
		netIn    []*prometheus.Desc
		netOut   []*prometheus.Desc
		uptime   []*prometheus.Desc
		maps     []*prometheus.Desc
		fps      []*prometheus.Desc
		players  []*prometheus.Desc
		connects []*prometheus.Desc
	)
	for _, server := range config.Targets {
		cpu = append(cpu, createStatsDesc("cpu", server.Name))
		netIn = append(cpu, createStatsDesc("net_in", server.Name))
		netOut = append(cpu, createStatsDesc("net_out", server.Name))
		uptime = append(cpu, createStatsDesc("uptime", server.Name))
		maps = append(cpu, createStatsDesc("maps", server.Name))
		fps = append(cpu, createStatsDesc("fps", server.Name))
		players = append(cpu, createStatsDesc("players", server.Name))
		connects = append(cpu, createStatsDesc("connects", server.Name))
	}
	return &statsCollector{
		config:   config,
		cpu:      cpu,
		netIn:    netIn,
		netOut:   netOut,
		uptime:   uptime,
		maps:     maps,
		fps:      fps,
		players:  players,
		connects: connects,
	}
}
func (s *statsCollector) Name() string {
	return "stats"
}
func (s *statsCollector) Describe(_ chan<- *prometheus.Desc) {
}

func (s *statsCollector) Update(ch chan<- prometheus.Metric, ctx context.Context) error {
	for _, server := range s.config.Targets {
		conn, errConn := cm.get(server)
		if errConn != nil {
			log.Error("Failed to get connection", zap.Error(errConn))
			continue
		}
		if connectErr := conn.Connect(ctx); connectErr != nil {
			log.Error("Failed to connect", zap.Error(errConn))
			continue
		}
		newStats, errStats := conn.Stats()
		if errStats != nil {
			log.Error("Failed to get stats", zap.Error(errStats))
			continue
		}
		if newStats == nil {
			log.Error("No stats returned")
			continue
		}
		cpu := createStatsDesc("cpu", server.Name)
		netIn := createStatsDesc("net_in", server.Name)
		netOut := createStatsDesc("net_out", server.Name)
		uptime := createStatsDesc("uptime", server.Name)
		maps := createStatsDesc("maps", server.Name)
		fps := createStatsDesc("fps", server.Name)
		players := createStatsDesc("players", server.Name)
		connects := createStatsDesc("connects", server.Name)

		ch <- prometheus.MustNewConstMetric(cpu, prometheus.GaugeValue, newStats.CPU)
		ch <- prometheus.MustNewConstMetric(netIn, prometheus.GaugeValue, newStats.NetIn)
		ch <- prometheus.MustNewConstMetric(netOut, prometheus.GaugeValue, newStats.NetOut)
		ch <- prometheus.MustNewConstMetric(uptime, prometheus.GaugeValue, float64(newStats.Uptime))
		ch <- prometheus.MustNewConstMetric(maps, prometheus.GaugeValue, float64(newStats.Maps))
		ch <- prometheus.MustNewConstMetric(fps, prometheus.GaugeValue, newStats.FPS)
		ch <- prometheus.MustNewConstMetric(players, prometheus.GaugeValue, float64(newStats.Player))
		ch <- prometheus.MustNewConstMetric(connects, prometheus.GaugeValue, float64(newStats.Connects))
	}
	return nil
}

func parseStats(body string) (*stats, error) {
	for _, line := range strings.Split(body, "\n") {
		match := reStats.FindStringSubmatch(line)
		if match != nil {
			return &stats{
				CPU:      toFloat64Default(match[1], 0.0),
				NetIn:    toFloat64Default(match[2], 0.0),
				NetOut:   toFloat64Default(match[3], 0.0),
				Uptime:   toIntDefault(match[4], 0),
				Maps:     toIntDefault(match[5], 0),
				FPS:      toFloat64Default(match[6], 0.0),
				Player:   toIntDefault(match[7], 0),
				Connects: toIntDefault(match[8], 0),
			}, nil
		}
	}
	return nil, errors.New("Failed to parse stats")
}

func init() {
	reStats = regexp.MustCompile(`^(?P<cpu>\d{1,3}\.\d{1,2})\s+(?P<net_in>\d{1,3}\.\d{1,2})\s+(?P<net_out>\d{1,3}\.\d{1,2})\s+(?P<uptime>\d+)\s+(?P<maps>\d+)\s+(?P<fps>\d{1,3}\.\d{1,2})\s+(?P<players>\d+)\s+(?P<connects>\d+)(\s+)?$`)
}
