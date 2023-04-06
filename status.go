package main

import (
	"context"
	"fmt"
	"github.com/leighmacdonald/steamid/v2/steamid"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	reMapName *regexp.Regexp
	rePlayers *regexp.Regexp
	rePlayer  *regexp.Regexp
	reEdicts  *regexp.Regexp
)

type statusPlayer struct {
	online  int
	ping    int
	loss    int
	address string
	port    int
	ip      string
	steamID steamid.SID64
}

type status struct {
	Map           string
	Players       []statusPlayer
	PlayerLimit   int
	PlayersHumans int
	PlayersBots   int
	Edicts        int
}

type statusCollector struct {
	config       *config
	connected    []*prometheus.Desc
	ping         []*prometheus.Desc
	loss         []*prometheus.Desc
	mapName      []*prometheus.Desc
	edicts       []*prometheus.Desc
	playersCount []*prometheus.Desc
	playersLimit []*prometheus.Desc
	playersHuman []*prometheus.Desc
	playersBots  []*prometheus.Desc
}

func createStatusDesc(stat string, labels prometheus.Labels) *prometheus.Desc {
	switch stat {
	case "edicts":
		return prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "status", stat),
			"The current edict usage (2048 max)",
			nil, labels)
	case "connected":
		return prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "status", stat),
			"The duration the player has been connected for in seconds",
			nil, labels)
	case "ping":
		return prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "status", stat),
			"The current player ping",
			nil, labels)
	case "loss":
		return prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "status", stat),
			"The current player loss",
			nil, labels)
	case "map_name":
		return prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "status", stat),
			"The current map name",
			nil, labels)
	case "players_count":
		return prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "status", stat),
			"The current server player count",
			nil, labels)
	case "players_limit":
		return prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "status", stat),
			"The current server player limit",
			nil, labels)
	case "players_human":
		return prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "status", stat),
			"The current server human player count",
			nil, labels)
	case "players_bots":
		return prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "status", stat),
			"The current server bot player limit",
			nil, labels)
	default:
		log.Panic("Unhandled stat Name", zap.String("stat", stat))
	}
	return nil
}

func newStatusCollector(config *config) *statusCollector {
	var (
		connected    []*prometheus.Desc
		ping         []*prometheus.Desc
		loss         []*prometheus.Desc
		mapName      []*prometheus.Desc
		edicts       []*prometheus.Desc
		playersCount []*prometheus.Desc
		playersLimit []*prometheus.Desc
		playersHuman []*prometheus.Desc
		playersBots  []*prometheus.Desc
	)
	for _, server := range config.Targets {
		labels := prometheus.Labels{"server": server.Name}
		connected = append(connected, createStatusDesc("connected", labels))
		ping = append(ping, createStatusDesc("ping", labels))
		loss = append(ping, createStatusDesc("loss", labels))
		mapName = append(mapName, createStatusDesc("map_name", labels))
		edicts = append(mapName, createStatusDesc("edicts", labels))
		playersCount = append(playersCount, createStatusDesc("players_count", labels))
		playersLimit = append(playersLimit, createStatusDesc("players_limit", labels))
		playersHuman = append(playersHuman, createStatusDesc("players_human", labels))
		playersBots = append(playersBots, createStatusDesc("players_bots", labels))
	}
	return &statusCollector{
		config:       config,
		connected:    connected,
		ping:         ping,
		loss:         loss,
		mapName:      mapName,
		edicts:       edicts,
		playersCount: playersCount,
		playersLimit: playersLimit,
		playersHuman: playersHuman,
		playersBots:  playersBots,
	}
}
func (s *statusCollector) Name() string {
	return "status"
}
func (s *statusCollector) Describe(_ chan<- *prometheus.Desc) {
}

func (s *statusCollector) Update(ctx context.Context, ch chan<- prometheus.Metric) error {
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
		newStatus, errStats := conn.Status()
		if errStats != nil {
			log.Error("Failed to get status", zap.Error(errStats))
			continue
		}
		if newStatus == nil {
			log.Error("No status returned")
			continue
		}
		for _, player := range newStatus.Players {
			connected := createStatusDesc("connected", prometheus.Labels{"server": server.Name, "steam_id": player.steamID.String()})
			ping := createStatusDesc("ping", prometheus.Labels{"server": server.Name, "steam_id": player.steamID.String()})
			loss := createStatusDesc("loss", prometheus.Labels{"server": server.Name, "steam_id": player.steamID.String()})
			ch <- prometheus.MustNewConstMetric(connected, prometheus.GaugeValue, float64(1))
			ch <- prometheus.MustNewConstMetric(ping, prometheus.GaugeValue, float64(player.ping))
			ch <- prometheus.MustNewConstMetric(loss, prometheus.GaugeValue, float64(player.loss))
		}

		playersCount := createStatusDesc("players_count", prometheus.Labels{"server": server.Name})
		playersLimit := createStatusDesc("players_limit", prometheus.Labels{"server": server.Name})
		playersHuman := createStatusDesc("players_human", prometheus.Labels{"server": server.Name})
		playersBots := createStatusDesc("players_bots", prometheus.Labels{"server": server.Name})
		edicts := createStatusDesc("edicts", prometheus.Labels{"server": server.Name})

		ch <- prometheus.MustNewConstMetric(playersCount, prometheus.GaugeValue, float64(len(newStatus.Players)))
		ch <- prometheus.MustNewConstMetric(playersLimit, prometheus.GaugeValue, float64(newStatus.PlayerLimit))
		ch <- prometheus.MustNewConstMetric(playersHuman, prometheus.GaugeValue, float64(newStatus.PlayersHumans))
		ch <- prometheus.MustNewConstMetric(playersBots, prometheus.GaugeValue, float64(newStatus.PlayersBots))
		ch <- prometheus.MustNewConstMetric(edicts, prometheus.GaugeValue, float64(newStatus.Edicts))
	}
	return nil
}
func parseConnected(d string) (time.Duration, error) {
	pcs := strings.Split(d, ":")
	var dur time.Duration
	var parseErr error
	switch len(pcs) {
	case 3:
		dur, parseErr = time.ParseDuration(fmt.Sprintf("%sh%sm%ss", pcs[0], pcs[1], pcs[2]))
	case 2:
		dur, parseErr = time.ParseDuration(fmt.Sprintf("%sm%ss", pcs[0], pcs[1]))
	case 1:
		dur, parseErr = time.ParseDuration(fmt.Sprintf("%ss", pcs[0]))
	default:
		dur = 0
	}
	return dur, parseErr
}
func parseStatus(body string) (*status, error) {
	s := status{}
	for _, line := range strings.Split(body, "\n") {
		match := reMapName.FindStringSubmatch(line)
		if match != nil {
			s.Map = match[1]
			continue
		}

		match = reEdicts.FindStringSubmatch(line)
		if match != nil {
			s.Edicts = toIntDefault(match[1], 0)
			continue
		}

		match = rePlayers.FindStringSubmatch(line)
		if match != nil {
			s.PlayersHumans = toIntDefault(match[1], 0)
			s.PlayersBots = toIntDefault(match[2], 0)
			s.PlayerLimit = toIntDefault(match[3], 32)
			continue
		}

		match = rePlayer.FindStringSubmatch(line)
		if match != nil {
			p := statusPlayer{}
			p.steamID = steamid.SID3ToSID64(steamid.SID3(match[3]))
			duration, errDur := parseConnected(match[4])
			if errDur != nil {
				log.Error("Failed to parse time connected", zap.Error(errDur))
				duration = 0
			}

			p.online = int(duration.Seconds())
			p.ping = toIntDefault(match[6], 0)
			p.loss = toIntDefault(match[7], 0)
			p.address = match[9]
			pcs := strings.Split(p.address, ":")
			p.ip = pcs[0]
			port, errPort := strconv.ParseInt(pcs[1], 10, 64)
			if errPort != nil {
				log.Error("Failed to parse port", zap.Error(errPort))
				port = 20000
			}
			p.port = int(port)
			s.Players = append(s.Players, p)
		}
	}

	return &s, nil
}

func init() {
	reMapName = regexp.MustCompile(`^map\s{5}:\s(?P<map_name>.+?)\sat.+?$`)
	reEdicts = regexp.MustCompile(`^edicts\s+:\s+(?P<edicts>\d+)\sused.+?$`)
	rePlayers = regexp.MustCompile(`^players\s+:\s+(?P<humans>\d+)\s+humans,\s+(?P<bots>\d+)\s+bots\s+\((?P<max>\d+)\smax\)$`)
	rePlayer = regexp.MustCompile(`^#\s{1,6}(?P<id>\d{1,6})\s"(?P<name>.+?)"\s+(?P<sid>\[U:\d:\d{1,10}])\s{1,8}(?P<time>\d{1,3}:\d{2}(:\d{2})?)\s+(?P<ping>\d{1,4})\s{1,8}(?P<loss>\d{1,3})\s(spawning|active)\s+(?P<ip>\d+\.\d+\.\d+\.\d+:\d+)$`)
}
