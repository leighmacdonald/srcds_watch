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
	"sync"
	"time"
)

var (
	reMapName        *regexp.Regexp
	rePlayers        *regexp.Regexp
	rePlayer         *regexp.Regexp
	reEdicts         *regexp.Regexp
	reVisiblePlayers *regexp.Regexp
	reStats          *regexp.Regexp
	reRate           *regexp.Regexp
	reMMVersion      *regexp.Regexp
	reSMVersion      *regexp.Regexp
	reSourceTV       *regexp.Regexp
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
	Map                 string
	Players             []statusPlayer
	PlayerLimit         int
	PlayersHumans       int
	PlayersBots         int
	Edicts              int
	SvVisibleMaxPlayers int
	SourceTV            bool
	CPU                 float64
	NetIn               float64
	NetOut              float64
	Uptime              int
	Maps                int
	FPS                 float64
	Player              int
	Connects            int
	SvMaXUpdateRate     float64
	SMVersion           string
	MMVersion           string
}

type statusCollector struct {
	config *config

	connected           []*prometheus.Desc
	online              []*prometheus.Desc
	svVisibleMaxPlayers []*prometheus.Desc
	ping                []*prometheus.Desc
	sourceTV            []*prometheus.Desc
	loss                []*prometheus.Desc
	mapName             []*prometheus.Desc
	edicts              []*prometheus.Desc
	playersCount        []*prometheus.Desc
	playersLimit        []*prometheus.Desc
	playersHuman        []*prometheus.Desc
	playersBots         []*prometheus.Desc
	cpu                 []*prometheus.Desc
	netIn               []*prometheus.Desc
	netOut              []*prometheus.Desc
	uptime              []*prometheus.Desc
	maps                []*prometheus.Desc
	fps                 []*prometheus.Desc
	players             []*prometheus.Desc
	connects            []*prometheus.Desc
	svMaxUpdateRate     []*prometheus.Desc
	smVersion           []*prometheus.Desc
	mmVersion           []*prometheus.Desc
}

func createStatusDesc(stat string, labels prometheus.Labels) *prometheus.Desc {
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
	case "sv_visiblemaxplayers":
		return prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "stats", stat),
			"The currently configured sv_visiblemaxplayers value",
			nil, labels)
	case "online":
		return prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "stats", stat),
			"1 if the game server is online",
			nil, labels)
	case "source_tv":
		return prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "stats", stat),
			"The current status of source tv",
			nil, labels)
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
		connected           []*prometheus.Desc
		online              []*prometheus.Desc
		sourceTV            []*prometheus.Desc
		svVisibleMaxPlayers []*prometheus.Desc
		ping                []*prometheus.Desc
		loss                []*prometheus.Desc
		mapName             []*prometheus.Desc
		edicts              []*prometheus.Desc
		playersCount        []*prometheus.Desc
		playersLimit        []*prometheus.Desc
		playersHuman        []*prometheus.Desc
		playersBots         []*prometheus.Desc
		cpu                 []*prometheus.Desc
		netIn               []*prometheus.Desc
		netOut              []*prometheus.Desc
		uptime              []*prometheus.Desc
		maps                []*prometheus.Desc
		fps                 []*prometheus.Desc
		players             []*prometheus.Desc
		connects            []*prometheus.Desc
		svMaxUpdateRate     []*prometheus.Desc
		mmVersion           []*prometheus.Desc
		smVersion           []*prometheus.Desc
	)
	for _, server := range config.Targets {
		labels := prometheus.Labels{"server": server.Name}
		online = append(connected, createStatusDesc("online", labels))
		connected = append(connected, createStatusDesc("connected", labels))
		sourceTV = append(sourceTV, createStatusDesc("source_tv", labels))
		svVisibleMaxPlayers = append(svVisibleMaxPlayers, createStatusDesc("sv_visiblemaxplayers", labels))
		ping = append(ping, createStatusDesc("ping", labels))
		loss = append(ping, createStatusDesc("loss", labels))
		mapName = append(mapName, createStatusDesc("map_name", labels))
		edicts = append(mapName, createStatusDesc("edicts", labels))
		playersCount = append(playersCount, createStatusDesc("players_count", labels))
		playersLimit = append(playersLimit, createStatusDesc("players_limit", labels))
		playersHuman = append(playersHuman, createStatusDesc("players_human", labels))
		playersBots = append(playersBots, createStatusDesc("players_bots", labels))
		cpu = append(cpu, createStatusDesc("cpu", labels))
		netIn = append(netIn, createStatusDesc("net_in", labels))
		netOut = append(netOut, createStatusDesc("net_out", labels))
		uptime = append(uptime, createStatusDesc("uptime", labels))
		maps = append(maps, createStatusDesc("maps", labels))
		fps = append(fps, createStatusDesc("fps", labels))
		players = append(players, createStatusDesc("players", labels))
		connects = append(connects, createStatusDesc("connects", labels))
		svMaxUpdateRate = append(svMaxUpdateRate, createStatusDesc("sv_max_update_rate", labels))
		mmVersion = append(mmVersion, createStatusDesc("metamod_version", labels))
		smVersion = append(smVersion, createStatusDesc("sourcemod_version", labels))
	}
	return &statusCollector{
		config:              config,
		cpu:                 cpu,
		netIn:               netIn,
		netOut:              netOut,
		uptime:              uptime,
		maps:                maps,
		fps:                 fps,
		players:             players,
		connects:            connects,
		svMaxUpdateRate:     svMaxUpdateRate,
		mmVersion:           mmVersion,
		smVersion:           smVersion,
		online:              online,
		connected:           connected,
		sourceTV:            sourceTV,
		svVisibleMaxPlayers: svVisibleMaxPlayers,
		ping:                ping,
		loss:                loss,
		mapName:             mapName,
		edicts:              edicts,
		playersCount:        playersCount,
		playersLimit:        playersLimit,
		playersHuman:        playersHuman,
		playersBots:         playersBots,
	}
}

func (s *statusCollector) Name() string {
	return "status"
}

func (s *statusCollector) Describe(_ chan<- *prometheus.Desc) {
}

func (s *statusCollector) Update(ctx context.Context, ch chan<- prometheus.Metric) error {
	wg := &sync.WaitGroup{}
	for _, target := range s.config.Targets {
		wg.Add(1)
		go func(server Target) {
			defer wg.Done()
			conn, errConn := cm.get(server)
			if errConn != nil {
				log.Error("Failed to get connection", zap.Error(errConn))
				return
			}
			if connectErr := conn.Connect(ctx); connectErr != nil {
				log.Error("Failed to connect", zap.Error(errConn))
				return
			}
			newStatus, errStats := conn.Status()
			if errStats != nil {
				log.Error("Failed to get status", zap.Error(errStats))
				return
			}
			if newStatus == nil {
				log.Error("No status returned")
				return
			}
			for _, player := range newStatus.Players {
				connected := createStatusDesc("connected", prometheus.Labels{"server": server.Name, "steam_id": player.steamID.String()})
				ping := createStatusDesc("ping", prometheus.Labels{"server": server.Name, "steam_id": player.steamID.String()})
				loss := createStatusDesc("loss", prometheus.Labels{"server": server.Name, "steam_id": player.steamID.String()})

				ch <- prometheus.MustNewConstMetric(connected, prometheus.GaugeValue, float64(1))
				ch <- prometheus.MustNewConstMetric(ping, prometheus.GaugeValue, float64(player.ping))
				ch <- prometheus.MustNewConstMetric(loss, prometheus.GaugeValue, float64(player.loss))
			}
			online := createStatusDesc("online", prometheus.Labels{"server": server.Name})
			playersCount := createStatusDesc("players_count", prometheus.Labels{"server": server.Name})
			playersLimit := createStatusDesc("players_limit", prometheus.Labels{"server": server.Name})
			playersHuman := createStatusDesc("players_human", prometheus.Labels{"server": server.Name})
			playersBots := createStatusDesc("players_bots", prometheus.Labels{"server": server.Name})
			edicts := createStatusDesc("edicts", prometheus.Labels{"server": server.Name})
			svVisibleMaxPlayers := createStatusDesc("sv_visiblemaxplayers", prometheus.Labels{"server": server.Name})
			sourceTV := createStatusDesc("source_tv", prometheus.Labels{"server": server.Name})
			cpu := createStatusDesc("cpu", prometheus.Labels{"server": server.Name})
			netIn := createStatusDesc("net_in", prometheus.Labels{"server": server.Name})
			netOut := createStatusDesc("net_out", prometheus.Labels{"server": server.Name})
			uptime := createStatusDesc("uptime", prometheus.Labels{"server": server.Name})
			maps := createStatusDesc("maps", prometheus.Labels{"server": server.Name})
			fps := createStatusDesc("fps", prometheus.Labels{"server": server.Name})
			players := createStatusDesc("players", prometheus.Labels{"server": server.Name})
			connects := createStatusDesc("connects", prometheus.Labels{"server": server.Name})
			svMaxUpdateRate := createStatusDesc("sv_max_update_rate", prometheus.Labels{"server": server.Name})
			mmVersion := createStatusDesc("metamod_version", prometheus.Labels{"server": server.Name,
				"metamod_version": newStatus.MMVersion})
			smVersion := createStatusDesc("sourcemod_version",
				prometheus.Labels{"server": server.Name, "sourcemod_version": newStatus.SMVersion})

			ch <- prometheus.MustNewConstMetric(online, prometheus.GaugeValue, 1)
			ch <- prometheus.MustNewConstMetric(playersCount, prometheus.GaugeValue, float64(len(newStatus.Players)))
			ch <- prometheus.MustNewConstMetric(playersLimit, prometheus.GaugeValue, float64(newStatus.PlayerLimit))
			ch <- prometheus.MustNewConstMetric(playersHuman, prometheus.GaugeValue, float64(newStatus.PlayersHumans))
			ch <- prometheus.MustNewConstMetric(playersBots, prometheus.GaugeValue, float64(newStatus.PlayersBots))
			ch <- prometheus.MustNewConstMetric(edicts, prometheus.GaugeValue, float64(newStatus.Edicts))
			ch <- prometheus.MustNewConstMetric(svVisibleMaxPlayers, prometheus.GaugeValue, float64(newStatus.SvVisibleMaxPlayers))
			if newStatus.SourceTV {
				ch <- prometheus.MustNewConstMetric(sourceTV, prometheus.GaugeValue, 1)
			} else {
				ch <- prometheus.MustNewConstMetric(sourceTV, prometheus.GaugeValue, 0)
			}
			ch <- prometheus.MustNewConstMetric(cpu, prometheus.GaugeValue, newStatus.CPU)
			ch <- prometheus.MustNewConstMetric(netIn, prometheus.GaugeValue, newStatus.NetIn)
			ch <- prometheus.MustNewConstMetric(netOut, prometheus.GaugeValue, newStatus.NetOut)
			ch <- prometheus.MustNewConstMetric(uptime, prometheus.GaugeValue, float64(newStatus.Uptime))
			ch <- prometheus.MustNewConstMetric(maps, prometheus.GaugeValue, float64(newStatus.Maps))
			ch <- prometheus.MustNewConstMetric(fps, prometheus.GaugeValue, newStatus.FPS)
			ch <- prometheus.MustNewConstMetric(players, prometheus.GaugeValue, float64(newStatus.Player))
			ch <- prometheus.MustNewConstMetric(connects, prometheus.GaugeValue, float64(newStatus.Connects))
			ch <- prometheus.MustNewConstMetric(svMaxUpdateRate, prometheus.GaugeValue, newStatus.SvMaXUpdateRate)
			ch <- prometheus.MustNewConstMetric(mmVersion, prometheus.GaugeValue, 1)
			ch <- prometheus.MustNewConstMetric(smVersion, prometheus.GaugeValue, 1)
		}(target)
	}
	wg.Wait()
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
		match = reVisiblePlayers.FindStringSubmatch(line)
		if match != nil {
			s.SvVisibleMaxPlayers = toIntDefault(match[1], 0)
			continue
		}
		match = reStats.FindStringSubmatch(line)
		if match != nil {
			s.CPU = toFloat64Default(match[1], 0.0)
			s.NetIn = toFloat64Default(match[2], 0.0)
			s.NetOut = toFloat64Default(match[3], 0.0)
			s.Uptime = toIntDefault(match[4], 0)
			s.Maps = toIntDefault(match[5], 0)
			s.FPS = toFloat64Default(match[6], 0.0)
			s.Player = toIntDefault(match[7], 0)
			s.Connects = toIntDefault(match[8], 0)
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
		match = reSourceTV.FindStringSubmatch(line)
		if match != nil {
			s.SourceTV = true
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
	reSourceTV = regexp.MustCompile(`^sourcetv:\s+(?P<stv>74.91.117.2:27015),`)
	reMMVersion = regexp.MustCompile(`^\s+Metamod:Source\sversion\s+(?P<mm_version>.+?)$`)
	reSMVersion = regexp.MustCompile(`^\s+SourceMod\sVersion:\s(?P<sm_version>.+?)$`)
	reRate = regexp.MustCompile(`^"sv_maxupdaterate" = "(?P<rate>\d+)"$`)
	reStats = regexp.MustCompile(`^(?P<cpu>\d{1,3}\.\d{1,2})\s+(?P<net_in>\d{1,3}\.\d{1,2})\s+(?P<net_out>\d{1,3}\.\d{1,2})\s+(?P<uptime>\d+)\s+(?P<maps>\d+)\s+(?P<fps>\d{1,3}\.\d{1,2})\s+(?P<players>\d+)\s+(?P<connects>\d+)(\s+)?$`)
	reVisiblePlayers = regexp.MustCompile(`^"sv_visiblemaxplayers" = "(?P<sv_visiblemaxplayers>\d+)"`)
	reMapName = regexp.MustCompile(`^map\s{5}:\s(?P<map_name>.+?)\sat.+?$`)
	reEdicts = regexp.MustCompile(`^edicts\s+:\s+(?P<edicts>\d+)\sused.+?$`)
	rePlayers = regexp.MustCompile(`^players\s+:\s+(?P<humans>\d+)\s+humans,\s+(?P<bots>\d+)\s+bots\s+\((?P<max>\d+)\smax\)$`)
	rePlayer = regexp.MustCompile(`^#\s{1,6}(?P<id>\d{1,6})\s"(?P<name>.+?)"\s+(?P<sid>\[U:\d:\d{1,10}])\s{1,8}(?P<time>\d{1,3}:\d{2}(:\d{2})?)\s+(?P<ping>\d{1,4})\s{1,8}(?P<loss>\d{1,3})\s(spawning|active)\s+(?P<ip>\d+\.\d+\.\d+\.\d+:\d+)$`)
}
