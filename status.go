package main

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/leighmacdonald/rcon/rcon"
	"github.com/leighmacdonald/steamid/v4/steamid"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

type statusPlayer struct {
	online  int
	ping    int
	loss    int
	address string
	port    int
	ip      string
	steamID steamid.SteamID
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

func createStatusDesc(namespace string, stat string, labels prometheus.Labels) *prometheus.Desc {
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
		slog.Warn("Unhandled stat Name", slog.String("stat", stat))
	}

	return nil
}

func newStatusCollector(config *config) *statusCollector {
	var ( //nolint:prealloc
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

		online = append(online, createStatusDesc(config.NameSpace, "online", labels))
		connected = append(connected, createStatusDesc(config.NameSpace, "connected", labels))
		sourceTV = append(sourceTV, createStatusDesc(config.NameSpace, "source_tv", labels))
		svVisibleMaxPlayers = append(svVisibleMaxPlayers, createStatusDesc(config.NameSpace, "sv_visiblemaxplayers", labels))
		ping = append(ping, createStatusDesc(config.NameSpace, "ping", labels))
		loss = append(loss, createStatusDesc(config.NameSpace, "loss", labels))
		mapName = append(mapName, createStatusDesc(config.NameSpace, "map_name", labels))
		edicts = append(edicts, createStatusDesc(config.NameSpace, "edicts", labels))
		playersCount = append(playersCount, createStatusDesc(config.NameSpace, "players_count", labels))
		playersLimit = append(playersLimit, createStatusDesc(config.NameSpace, "players_limit", labels))
		playersHuman = append(playersHuman, createStatusDesc(config.NameSpace, "players_human", labels))
		playersBots = append(playersBots, createStatusDesc(config.NameSpace, "players_bots", labels))
		cpu = append(cpu, createStatusDesc(config.NameSpace, "cpu", labels))
		netIn = append(netIn, createStatusDesc(config.NameSpace, "net_in", labels))
		netOut = append(netOut, createStatusDesc(config.NameSpace, "net_out", labels))
		uptime = append(uptime, createStatusDesc(config.NameSpace, "uptime", labels))
		maps = append(maps, createStatusDesc(config.NameSpace, "maps", labels))
		fps = append(fps, createStatusDesc(config.NameSpace, "fps", labels))
		players = append(players, createStatusDesc(config.NameSpace, "players", labels))
		connects = append(connects, createStatusDesc(config.NameSpace, "connects", labels))
		svMaxUpdateRate = append(svMaxUpdateRate, createStatusDesc(config.NameSpace, "sv_max_update_rate", labels))
		mmVersion = append(mmVersion, createStatusDesc(config.NameSpace, "metamod_version", labels))
		smVersion = append(smVersion, createStatusDesc(config.NameSpace, "sourcemod_version", labels))
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

func (s *statusCollector) Update(ctx context.Context, metricCHan chan<- prometheus.Metric) error {
	waitGroup := &sync.WaitGroup{}

	for _, target := range s.config.Targets {
		waitGroup.Add(1)

		go func(server Target) {
			defer waitGroup.Done()

			conn, errConn := rcon.Dial(ctx, server.addr(), server.Password, time.Second*8)
			if errConn != nil {
				slog.Error("Failed to connect", slog.String("server", server.Name), slog.String("error", errConn.Error()))

				return
			}

			defer func() {
				if err := conn.Close(); err != nil {
					slog.Error("Failed to close connection", slog.String("server", server.Name), slog.String("error", errConn.Error()))
				}
			}()

			newStatus, errStats := fetchStatus(conn)
			if errStats != nil {
				slog.Error("Failed to get status", slog.String("server", server.Name), slog.String("error", errStats.Error()))

				return
			}

			if newStatus == nil {
				slog.Error("No status returned", slog.String("server", server.Name))

				return
			}

			slog.Debug("Got status", slog.String("map", newStatus.Map), slog.String("server", server.Name))

			for _, player := range newStatus.Players {
				connected := createStatusDesc(s.config.NameSpace, "connected", prometheus.Labels{"server": server.Name, "steam_id": player.steamID.String()})
				ping := createStatusDesc(s.config.NameSpace, "ping", prometheus.Labels{"server": server.Name, "steam_id": player.steamID.String()})
				loss := createStatusDesc(s.config.NameSpace, "loss", prometheus.Labels{"server": server.Name, "steam_id": player.steamID.String()})

				metricCHan <- prometheus.MustNewConstMetric(connected, prometheus.GaugeValue, float64(1))
				metricCHan <- prometheus.MustNewConstMetric(ping, prometheus.GaugeValue, float64(player.ping))
				metricCHan <- prometheus.MustNewConstMetric(loss, prometheus.GaugeValue, float64(player.loss))
			}

			online := createStatusDesc(s.config.NameSpace, "online", prometheus.Labels{"server": server.Name})
			playersCount := createStatusDesc(s.config.NameSpace, "players_count", prometheus.Labels{"server": server.Name})
			playersLimit := createStatusDesc(s.config.NameSpace, "players_limit", prometheus.Labels{"server": server.Name})
			playersHuman := createStatusDesc(s.config.NameSpace, "players_human", prometheus.Labels{"server": server.Name})
			playersBots := createStatusDesc(s.config.NameSpace, "players_bots", prometheus.Labels{"server": server.Name})
			edicts := createStatusDesc(s.config.NameSpace, "edicts", prometheus.Labels{"server": server.Name})
			svVisibleMaxPlayers := createStatusDesc(s.config.NameSpace, "sv_visiblemaxplayers", prometheus.Labels{"server": server.Name})
			sourceTV := createStatusDesc(s.config.NameSpace, "source_tv", prometheus.Labels{"server": server.Name})
			cpu := createStatusDesc(s.config.NameSpace, "cpu", prometheus.Labels{"server": server.Name})
			netIn := createStatusDesc(s.config.NameSpace, "net_in", prometheus.Labels{"server": server.Name})
			netOut := createStatusDesc(s.config.NameSpace, "net_out", prometheus.Labels{"server": server.Name})
			uptime := createStatusDesc(s.config.NameSpace, "uptime", prometheus.Labels{"server": server.Name})
			maps := createStatusDesc(s.config.NameSpace, "maps", prometheus.Labels{"server": server.Name})
			fps := createStatusDesc(s.config.NameSpace, "fps", prometheus.Labels{"server": server.Name})
			players := createStatusDesc(s.config.NameSpace, "players", prometheus.Labels{"server": server.Name})
			connects := createStatusDesc(s.config.NameSpace, "connects", prometheus.Labels{"server": server.Name})
			svMaxUpdateRate := createStatusDesc(s.config.NameSpace, "sv_max_update_rate", prometheus.Labels{"server": server.Name})
			mmVersion := createStatusDesc(s.config.NameSpace, "metamod_version", prometheus.Labels{
				"server":          server.Name,
				"metamod_version": newStatus.MMVersion,
			})
			smVersion := createStatusDesc(s.config.NameSpace, "sourcemod_version",
				prometheus.Labels{"server": server.Name, "sourcemod_version": newStatus.SMVersion})

			metricCHan <- prometheus.MustNewConstMetric(online, prometheus.GaugeValue, 1)
			metricCHan <- prometheus.MustNewConstMetric(playersCount, prometheus.GaugeValue, float64(len(newStatus.Players)))
			metricCHan <- prometheus.MustNewConstMetric(playersLimit, prometheus.GaugeValue, float64(newStatus.PlayerLimit))
			metricCHan <- prometheus.MustNewConstMetric(playersHuman, prometheus.GaugeValue, float64(newStatus.PlayersHumans))
			metricCHan <- prometheus.MustNewConstMetric(playersBots, prometheus.GaugeValue, float64(newStatus.PlayersBots))
			metricCHan <- prometheus.MustNewConstMetric(edicts, prometheus.GaugeValue, float64(newStatus.Edicts))
			metricCHan <- prometheus.MustNewConstMetric(svVisibleMaxPlayers, prometheus.GaugeValue, float64(newStatus.SvVisibleMaxPlayers))

			if newStatus.SourceTV {
				metricCHan <- prometheus.MustNewConstMetric(sourceTV, prometheus.GaugeValue, 1)
			} else {
				metricCHan <- prometheus.MustNewConstMetric(sourceTV, prometheus.GaugeValue, 0)
			}

			metricCHan <- prometheus.MustNewConstMetric(cpu, prometheus.GaugeValue, newStatus.CPU)
			metricCHan <- prometheus.MustNewConstMetric(netIn, prometheus.GaugeValue, newStatus.NetIn)
			metricCHan <- prometheus.MustNewConstMetric(netOut, prometheus.GaugeValue, newStatus.NetOut)
			metricCHan <- prometheus.MustNewConstMetric(uptime, prometheus.GaugeValue, float64(newStatus.Uptime))
			metricCHan <- prometheus.MustNewConstMetric(maps, prometheus.GaugeValue, float64(newStatus.Maps))
			metricCHan <- prometheus.MustNewConstMetric(fps, prometheus.GaugeValue, newStatus.FPS)
			metricCHan <- prometheus.MustNewConstMetric(players, prometheus.GaugeValue, float64(newStatus.Player))
			metricCHan <- prometheus.MustNewConstMetric(connects, prometheus.GaugeValue, float64(newStatus.Connects))
			metricCHan <- prometheus.MustNewConstMetric(svMaxUpdateRate, prometheus.GaugeValue, newStatus.SvMaXUpdateRate)
			metricCHan <- prometheus.MustNewConstMetric(mmVersion, prometheus.GaugeValue, 1)
			metricCHan <- prometheus.MustNewConstMetric(smVersion, prometheus.GaugeValue, 1)
		}(target)
	}

	waitGroup.Wait()

	return nil
}

func parseConnected(d string) (time.Duration, error) {
	var (
		pcs      = strings.Split(d, ":")
		dur      time.Duration
		parseErr error
	)

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

	return dur, errors.Wrap(parseErr, "Failed to parse connected time string")
}

func fetchStatus(conn *rcon.RemoteConsole) (*status, error) {
	body, errExec := conn.Exec("status;stats;sv_maxupdaterate;sm version;meta version;sv_visiblemaxplayers")

	if errExec != nil {
		return nil, errors.Wrap(errExec, "Failed to execute rcon status command")
	}

	parser := newStatusParser()

	return parser.parse(body)
}

type statusParser struct {
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
}

func newStatusParser() statusParser {
	return statusParser{
		reSourceTV:       regexp.MustCompile(`^sourcetv:\s+(?P<stv>74.91.117.2:27015),`),
		reMMVersion:      regexp.MustCompile(`^\s+Metamod:Source\sversion\s+(?P<mm_version>.+?)$`),
		reSMVersion:      regexp.MustCompile(`^\s+SourceMod\sVersion:\s(?P<sm_version>.+?)$`),
		reRate:           regexp.MustCompile(`^"sv_maxupdaterate" = "(?P<rate>\d+)"$`),
		reStats:          regexp.MustCompile(`^(?P<cpu>\d{1,3}\.\d{1,2})\s+(?P<net_in>\d{1,3}\.\d{1,2})\s+(?P<net_out>\d{1,3}\.\d{1,2})\s+(?P<uptime>\d+)\s+(?P<maps>\d+)\s+(?P<fps>\d{1,3}\.\d{1,2})\s+(?P<players>\d+)\s+(?P<connects>\d+)(\s+)?$`),
		reVisiblePlayers: regexp.MustCompile(`^"sv_visiblemaxplayers" = "(?P<sv_visiblemaxplayers>\d+)"`),
		reMapName:        regexp.MustCompile(`^map\s{5}:\s(?P<map_name>.+?)\sat.+?$`),
		reEdicts:         regexp.MustCompile(`^edicts\s+:\s+(?P<edicts>\d+)\sused.+?$`),
		rePlayers:        regexp.MustCompile(`^players\s+:\s+(?P<humans>\d+)\s+humans,\s+(?P<bots>\d+)\s+bots\s+\((?P<max>\d+)\smax\)$`),
		rePlayer:         regexp.MustCompile(`^#\s{1,6}(?P<id>\d{1,6})\s"(?P<name>.+?)"\s+(?P<sid>\[U:\d:\d{1,10}])\s{1,8}(?P<time>\d{1,3}:\d{2}(:\d{2})?)\s+(?P<ping>\d{1,4})\s{1,8}(?P<loss>\d{1,3})\s(spawning|active)\s+(?P<ip>\d+\.\d+\.\d+\.\d+:\d+)$`),
	}
}

func (p *statusParser) parse(body string) (*status, error) {
	newStatus := status{}

	for _, line := range strings.Split(body, "\n") {
		match := p.reMapName.FindStringSubmatch(line)
		if match != nil {
			newStatus.Map = match[1]

			continue
		}

		match = p.reEdicts.FindStringSubmatch(line)
		if match != nil {
			newStatus.Edicts = toIntDefault(match[1], 0)

			continue
		}

		match = p.rePlayers.FindStringSubmatch(line)
		if match != nil {
			newStatus.PlayersHumans = toIntDefault(match[1], 0)
			newStatus.PlayersBots = toIntDefault(match[2], 0)
			newStatus.PlayerLimit = toIntDefault(match[3], 32)

			continue
		}

		match = p.reVisiblePlayers.FindStringSubmatch(line)
		if match != nil {
			newStatus.SvVisibleMaxPlayers = toIntDefault(match[1], 0)

			continue
		}

		match = p.reStats.FindStringSubmatch(line)
		if match != nil {
			newStatus.CPU = toFloat64Default(match[1], 0.0)
			newStatus.NetIn = toFloat64Default(match[2], 0.0)
			newStatus.NetOut = toFloat64Default(match[3], 0.0)
			newStatus.Uptime = toIntDefault(match[4], 0)
			newStatus.Maps = toIntDefault(match[5], 0)
			newStatus.FPS = toFloat64Default(match[6], 0.0)
			newStatus.Player = toIntDefault(match[7], 0)
			newStatus.Connects = toIntDefault(match[8], 0)

			continue
		}

		match = p.reMMVersion.FindStringSubmatch(line)
		if match != nil {
			newStatus.MMVersion = match[1]

			continue
		}

		match = p.reSMVersion.FindStringSubmatch(line)
		if match != nil {
			newStatus.SMVersion = match[1]

			continue
		}

		match = p.reRate.FindStringSubmatch(line)
		if match != nil {
			newStatus.SvMaXUpdateRate = toFloat64Default(match[1], 0)

			continue
		}

		match = p.reSourceTV.FindStringSubmatch(line)
		if match != nil {
			newStatus.SourceTV = true

			continue
		}

		match = p.rePlayer.FindStringSubmatch(line)
		if match != nil {
			newStatusPlayer := statusPlayer{}
			newStatusPlayer.steamID = steamid.New(match[3])

			duration, errDur := parseConnected(match[4])
			if errDur != nil {
				duration = time.Duration(0)
			}

			newStatusPlayer.online = int(duration.Seconds())
			newStatusPlayer.ping = toIntDefault(match[6], 0)
			newStatusPlayer.loss = toIntDefault(match[7], 0)
			newStatusPlayer.address = match[9]
			pcs := strings.Split(newStatusPlayer.address, ":")
			newStatusPlayer.ip = pcs[0]

			port, errPort := strconv.ParseInt(pcs[1], 10, 64)
			if errPort != nil {
				port = 20000
			}

			newStatusPlayer.port = int(port)
			newStatus.Players = append(newStatus.Players, newStatusPlayer)
		}
	}

	return &newStatus, nil
}

func init() {
}
