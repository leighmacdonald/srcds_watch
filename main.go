package main

import (
	"context"
	"go.uber.org/zap"
	"os"
)

var (
	// Build info
	version = "master"
	commit  = "latest"
	date    = "n/a"
	builtBy = "src"

	log *zap.Logger
	cm  *connManager
)

const (
	namespace         = "srcds"
	defaultConfigPath = "srcds_watch.yml"
)

func main() {
	ctx := context.Background()
	vi := versionInfo{version: version, commit: commit, date: date, builtBy: builtBy}
	log = mustCreateLogger()
	log.Info("Starting srcds_watch",
		zap.String("version", vi.version),
		zap.String("commit", vi.commit),
		zap.String("date", vi.date))
	conf := newConfig()
	if !exists(defaultConfigPath) {
		log.Panic("Failed to find config file", zap.String("config_path", defaultConfigPath))
	}
	log.Info("Loading config file", zap.String("config_path", defaultConfigPath))
	cf, cfErr := os.Open(defaultConfigPath)
	if cfErr != nil {
		log.Panic("Failed to open config file", zap.Error(cfErr))
	}
	if errRead := conf.read(cf); errRead != nil {
		log.Panic("Failed to read config file", zap.Error(errRead))
	}
	if errClose := cf.Close(); errClose != nil {
		log.Error("Failed to close config file after read", zap.Error(errClose))
	}
	app := newApplication(conf)
	if errApp := app.start(ctx); errApp != nil {
		log.Error("Application returned error", zap.Error(errApp))
	}
}

func init() {
	cm = newConnManager()
}
