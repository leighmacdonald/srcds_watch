package main

import (
	"fmt"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"io"
	"os"
)

type versionInfo struct {
	version string
	commit  string
	date    string
	builtBy string
}

func exists(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}
	return true
}

// Target is the remote server configuration
type Target struct {
	Host     string `yaml:"host"`
	Port     uint16 `yaml:"port"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
}

func (t Target) addr() string {
	return fmt.Sprintf("%s:%d", t.Host, t.Port)
}

type config struct {
	ListenHost  string   `yaml:"listen_host"`
	ListenPort  uint16   `yaml:"listen_port"`
	MetricsPath string   `yaml:"metrics_path"`
	Targets     []Target `yaml:"targets"`
}

func (c *config) Addr() string {
	return fmt.Sprintf("%s:%d", c.ListenHost, c.ListenPort)
}

func newConfig() *config {
	return &config{
		ListenHost:  "0.0.0.0",
		ListenPort:  8767,
		MetricsPath: "/metrics",
		Targets:     nil,
	}
}

func (c *config) read(reader io.Reader) error {
	return yaml.NewDecoder(reader).Decode(c)
}

func mustCreateLogger() *zap.Logger {
	loggingConfig := zap.NewProductionConfig()
	//loggingConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logger, errLogger := loggingConfig.Build()
	if errLogger != nil {
		fmt.Printf("Failed to create logger: %v\n", errLogger)
		os.Exit(1)
	}
	return logger
}
