package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
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

func main() {
	ctx := context.Background()
	build := versionInfo{version: version, commit: commit, date: date, builtBy: builtBy}

	signalCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	log := mustCreateLogger()
	log.Info("Starting srcds_watch",
		zap.String("version", build.version),
		zap.String("commit", build.commit),
		zap.String("date", build.date))

	conf := newConfig()

	if !exists(defaultConfigPath) {
		log.Panic("Failed to find config file", zap.String("config_path", defaultConfigPath))
	}

	log.Info("Loading config file", zap.String("config_path", defaultConfigPath))

	configFile, cfErr := os.Open(defaultConfigPath)
	if cfErr != nil {
		log.Panic("Failed to open config file", zap.Error(cfErr))
	}

	if errRead := conf.read(configFile); errRead != nil {
		log.Panic("Failed to read config file", zap.Error(errRead))
	}

	if errClose := configFile.Close(); errClose != nil {
		log.Error("Failed to close config file after read", zap.Error(errClose))
	}

	app := newApplication(conf, log)
	if errApp := app.start(signalCtx); errApp != nil {
		log.Error("Application returned error", zap.Error(errApp))
	}

	<-signalCtx.Done()
}
