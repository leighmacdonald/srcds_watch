package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/dotse/slug"
)

var (
	// Build info.
	version = "master"
	commit  = "latest" //nolint:gochecknoglobals
	date    = "n/a"    //nolint:gochecknoglobals
	builtBy = "src"    //nolint:gochecknoglobals
)

const (
	defaultConfigPath = "srcds_watch.yml"
)

func mustCreateLogger(levelString string) func() {
	var level slog.Level

	switch levelString {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	default:
		level = slog.LevelError
	}

	closer := func() {}

	opts := slug.HandlerOptions{ //nolint:exhaustruct

		HandlerOptions: slog.HandlerOptions{ //nolint:exhaustruct

			Level: level,

			AddSource: false,
		},
	}

	slog.SetDefault(slog.New(slug.NewHandler(opts, os.Stdout)))

	return closer
}

func run() int {
	ctx := context.Background()
	build := versionInfo{version: version, commit: commit, date: date, builtBy: builtBy}

	signalCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	conf := newConfig()

	if !exists(defaultConfigPath) {
		slog.Error("Failed to find config file", slog.String("config_path", defaultConfigPath))

		return 1
	}

	configFile, cfErr := os.Open(defaultConfigPath)
	if cfErr != nil {
		slog.Error("Failed to open config file", slog.String("error", cfErr.Error()))

		return 1
	}

	if errRead := conf.read(configFile); errRead != nil {
		slog.Error("Failed to read config file", slog.String("error", errRead.Error()))

		return 1
	}

	if errClose := configFile.Close(); errClose != nil {
		slog.Error("Failed to close config file after read", slog.String("error", errClose.Error()))

		return 1
	}

	closeLog := mustCreateLogger(conf.LogLevel)
	defer closeLog()

	slog.Debug("Using config file", slog.String("config_path", defaultConfigPath))

	slog.Info("Starting srcds_watch",
		slog.String("version", build.version),
		slog.String("commit", build.commit),
		slog.String("date", build.date))

	if errApp := start(signalCtx, conf); errApp != nil {
		slog.Error("Application returned error", slog.String("error", errApp.Error()))

		return 1
	}

	<-signalCtx.Done()

	return 0
}

func main() {
	os.Exit(run())
}
