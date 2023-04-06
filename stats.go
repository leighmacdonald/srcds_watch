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
	reStats     *regexp.Regexp
	reRate      *regexp.Regexp
	reMMVersion *regexp.Regexp
	reSMVersion *regexp.Regexp
)

type stats struct {
	CPU             float64
	NetIn           float64
	NetOut          float64
	Uptime          int
	Maps            int
	FPS             float64
	Player          int
	Connects        int
	SvMaXUpdateRate float64
	SMVersion       string
	MMVersion       string
}

type statsCollector struct {
	config          *config
	cpu             []*prometheus.Desc
	netIn           []*prometheus.Desc
	netOut          []*prometheus.Desc
	uptime          []*prometheus.Desc
	maps            []*prometheus.Desc
	fps             []*prometheus.Desc
	players         []*prometheus.Desc
	connects        []*prometheus.Desc
	svMaxUpdateRate []*prometheus.Desc
	smVersion       []*prometheus.Desc
	mmVersion       []*prometheus.Desc
}

func createStatsDesc(stat string, labels prometheus.Labels) *prometheus.Desc {
	switch stat {
	case "cpu":
		return prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "stats", stat),
			"The current cpu usage.",
			nil, labels)
	case "net_in":
		return prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "stats", stat),
			"The current inbound network traffic rate (KB/s)",
			nil, labels)
	case "net_out":
		return prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "stats", stat),
			"The current outbound network traffic rate (KB/s)",
			nil, labels)
	case "uptime":
		return prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "stats", stat),
			"The current server uptime in minutes",
			nil, labels)
	case "maps":
		return prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "stats", stat),
			"The total number of maps that have been played",
			nil, labels)
	case "fps":
		return prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "stats", stat),
			"The current server fps (tickrate)",
			nil, labels)
	case "players":
		return prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "stats", stat),
			"The current statusPlayer count of the server.",
			nil, labels)
	case "connects":
		return prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "stats", stat),
			"The total number of players that have connected to the server.",
			nil, labels)
	case "sv_max_update_rate":
		return prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "stats", stat),
			"The time in MS per tick",
			nil, labels)
	case "sourcemod_version":
		return prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "stats", stat),
			"Current currently running sourcemod version",
			nil, labels)
	case "metamod_version":
		return prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "stats", stat),
			"Current currently running metamod version",
			nil, labels)
	default:
		log.Panic("Unhandled stat Name", zap.String("stat", stat))
	}
	return nil
}

func newStatsCollector(config *config) *statsCollector {
	var (
		cpu             []*prometheus.Desc
		netIn           []*prometheus.Desc
		netOut          []*prometheus.Desc
		uptime          []*prometheus.Desc
		maps            []*prometheus.Desc
		fps             []*prometheus.Desc
		players         []*prometheus.Desc
		connects        []*prometheus.Desc
		svMaxUpdateRate []*prometheus.Desc
		mmVersion       []*prometheus.Desc
		smVersion       []*prometheus.Desc
	)
	for _, server := range config.Targets {
		labels := prometheus.Labels{"server": server.Name}
		cpu = append(cpu, createStatsDesc("cpu", labels))
		netIn = append(netIn, createStatsDesc("net_in", labels))
		netOut = append(netOut, createStatsDesc("net_out", labels))
		uptime = append(uptime, createStatsDesc("uptime", labels))
		maps = append(maps, createStatsDesc("maps", labels))
		fps = append(fps, createStatsDesc("fps", labels))
		players = append(players, createStatsDesc("players", labels))
		connects = append(connects, createStatsDesc("connects", labels))
		svMaxUpdateRate = append(svMaxUpdateRate, createStatsDesc("sv_max_update_rate", labels))
		mmVersion = append(mmVersion, createStatsDesc("metamod_version", labels))
		smVersion = append(smVersion, createStatsDesc("sourcemod_version", labels))
	}
	return &statsCollector{
		config:          config,
		cpu:             cpu,
		netIn:           netIn,
		netOut:          netOut,
		uptime:          uptime,
		maps:            maps,
		fps:             fps,
		players:         players,
		connects:        connects,
		svMaxUpdateRate: svMaxUpdateRate,
		mmVersion:       mmVersion,
		smVersion:       smVersion,
	}
}
func (s *statsCollector) Name() string {
	return "stats"
}
func (s *statsCollector) Describe(_ chan<- *prometheus.Desc) {
}

func (s *statsCollector) Update(ctx context.Context, ch chan<- prometheus.Metric) error {
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
		cpu := createStatsDesc("cpu", prometheus.Labels{"server": server.Name})
		netIn := createStatsDesc("net_in", prometheus.Labels{"server": server.Name})
		netOut := createStatsDesc("net_out", prometheus.Labels{"server": server.Name})
		uptime := createStatsDesc("uptime", prometheus.Labels{"server": server.Name})
		maps := createStatsDesc("maps", prometheus.Labels{"server": server.Name})
		fps := createStatsDesc("fps", prometheus.Labels{"server": server.Name})
		players := createStatsDesc("players", prometheus.Labels{"server": server.Name})
		connects := createStatsDesc("connects", prometheus.Labels{"server": server.Name})
		svMaxUpdateRate := createStatsDesc("sv_max_update_rate", prometheus.Labels{"server": server.Name})
		mmVersion := createStatsDesc("metamod_version", prometheus.Labels{"server": server.Name,
			"metamod_version": newStats.MMVersion})
		smVersion := createStatsDesc("sourcemod_version",
			prometheus.Labels{"server": server.Name, "sourcemod_version": newStats.SMVersion})

		ch <- prometheus.MustNewConstMetric(cpu, prometheus.GaugeValue, newStats.CPU)
		ch <- prometheus.MustNewConstMetric(netIn, prometheus.GaugeValue, newStats.NetIn)
		ch <- prometheus.MustNewConstMetric(netOut, prometheus.GaugeValue, newStats.NetOut)
		ch <- prometheus.MustNewConstMetric(uptime, prometheus.GaugeValue, float64(newStats.Uptime))
		ch <- prometheus.MustNewConstMetric(maps, prometheus.GaugeValue, float64(newStats.Maps))
		ch <- prometheus.MustNewConstMetric(fps, prometheus.GaugeValue, newStats.FPS)
		ch <- prometheus.MustNewConstMetric(players, prometheus.GaugeValue, float64(newStats.Player))
		ch <- prometheus.MustNewConstMetric(connects, prometheus.GaugeValue, float64(newStats.Connects))
		ch <- prometheus.MustNewConstMetric(svMaxUpdateRate, prometheus.GaugeValue, newStats.SvMaXUpdateRate)
		ch <- prometheus.MustNewConstMetric(mmVersion, prometheus.GaugeValue, 1)
		ch <- prometheus.MustNewConstMetric(smVersion, prometheus.GaugeValue, 1)
	}
	return nil
}

func parseStats(body string) (*stats, error) {
	ok := false
	var s stats
	for _, line := range strings.Split(body, "\n") {
		match := reStats.FindStringSubmatch(line)
		if match != nil {
			s.CPU = toFloat64Default(match[1], 0.0)
			s.NetIn = toFloat64Default(match[2], 0.0)
			s.NetOut = toFloat64Default(match[3], 0.0)
			s.Uptime = toIntDefault(match[4], 0)
			s.Maps = toIntDefault(match[5], 0)
			s.FPS = toFloat64Default(match[6], 0.0)
			s.Player = toIntDefault(match[7], 0)
			s.Connects = toIntDefault(match[8], 0)
			ok = true
			continue
		}
		match = reMMVersion.FindStringSubmatch(line)
		if match != nil {
			s.MMVersion = match[1]
			continue
		}
		match = reSMVersion.FindStringSubmatch(line)
		if match != nil {
			s.SMVersion = match[1]
			continue
		}
		match = reRate.FindStringSubmatch(line)
		if match != nil {
			s.SvMaXUpdateRate = toFloat64Default(match[1], 0)
			continue
		}
	}
	if !ok {
		return nil, errors.New("Failed to parse stats")
	}
	return &s, nil
}

func init() {
	reMMVersion = regexp.MustCompile(`^\s+Metamod:Source\sversion\s+(?P<mm_version>.+?)$`)
	reSMVersion = regexp.MustCompile(`^\s+SourceMod\sVersion:\s(?P<sm_version>.+?)$`)
	reRate = regexp.MustCompile(`^"sv_maxupdaterate" = "(?P<rate>\d+)"$`)
	reStats = regexp.MustCompile(`^(?P<cpu>\d{1,3}\.\d{1,2})\s+(?P<net_in>\d{1,3}\.\d{1,2})\s+(?P<net_out>\d{1,3}\.\d{1,2})\s+(?P<uptime>\d+)\s+(?P<maps>\d+)\s+(?P<fps>\d{1,3}\.\d{1,2})\s+(?P<players>\d+)\s+(?P<connects>\d+)(\s+)?$`)
}
