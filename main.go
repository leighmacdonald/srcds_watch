package main

import (
	"context"
	"go.uber.org/zap"
	"os"
)

var (
	// Build info
	version string = "master"
	commit  string = "latest"
	date    string = "n/a"
	builtBy string = "src"

	log *zap.Logger
	cm  *connManager
)

const (
	namespace         = "srcds"
	defaultConfigPath = "srcds_watch.yml"
)

func main() {
	ctx := context.Background()
	versionInfo := Version{Version: version, Commit: commit, Date: date, BuiltBy: builtBy}
	log = mustCreateLogger()
	log.Info("Starting srcds_watch",
		zap.String("version", versionInfo.Version),
		zap.String("commit", versionInfo.Commit),
		zap.String("date", versionInfo.Date))
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
