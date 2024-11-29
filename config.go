package main

import (
	"fmt"
	"io"
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
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

// Target is the remote server configuration.
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
	LogLevel    string   `yaml:"log_level"`
	MetricsPath string   `yaml:"metrics_path"`
	NameSpace   string   `yaml:"name_space"`
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
		NameSpace:   "srcds",
	}
}

func (c *config) read(reader io.Reader) error {
	if err := yaml.NewDecoder(reader).Decode(c); err != nil {
		return errors.Wrap(err, "Could not decode config")
	}

	if c.NameSpace == "" {
		c.NameSpace = "srcds"
	}

	if c.MetricsPath == "" {
		c.MetricsPath = "/metrics"
	}

	if c.ListenHost == "" {
		c.ListenHost = "0.0.0.0"
	}

	if c.ListenPort == 0 {
		c.ListenPort = 8767
	}

	if c.LogLevel == "" {
		c.LogLevel = "info"
	}

	return nil
}
