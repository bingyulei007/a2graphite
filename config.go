package main

import (
	"fmt"
	"a2graphite/ceilometer"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

type logConfig struct {
	Filename string `yaml:"filename"`
	Level    string `yaml:"level"`
}

type graphiteConfig struct {
	URL            string        `yaml:"url"`
	ReconnectDelay time.Duration `yaml:"reconnect_delay"`
	Prefix         string        `yaml:"prefix"`
	BufferSize     int           `yaml:"buffer_size"`
}

type statsConfig struct {
	Enabled        bool          `yaml:"enabled"`
	Interval       time.Duration `yaml:"interval"`
	URL            string        `yaml:"url"`
	ReconnectDelay time.Duration `yaml:"reconnect_delay"`
	Prefix         string        `yaml:"prefix"`
	BufferSize     int           `yaml:"buffer_size"`
}

type profilerConfig struct {
	Enabled    bool   `yaml:"enabled"`
	ListenAddr string `yaml:"listen_addr"`
}

// Config is a2graphite main config
type Config struct {
	Log      *logConfig      `yaml:"log"`
	Graphite *graphiteConfig `yaml:"graphite"`
	Stats    *statsConfig    `yaml:"stats"`
	Profiler *profilerConfig `yaml:"profiler"`

	// Receiver configs
	Ceilometer *ceilometer.Config `yaml:"ceilometer"`
}

// NewConfig returns a default config instance
func NewConfig() *Config {
	config := &Config{
		Log: &logConfig{
			Filename: "/dev/stdout",
			Level:    "info",
		},
		Graphite: &graphiteConfig{
			ReconnectDelay: 100 * time.Microsecond,
			BufferSize:     100,
			Prefix:         "a2graphite.",
		},
		Stats: &statsConfig{
			Enabled:        true,
			Interval:       60 * time.Second,
			ReconnectDelay: 100 * time.Microsecond,
			BufferSize:     100,
			Prefix:         "a2graphite-stats.",
		},
		Profiler: &profilerConfig{
			Enabled:    false,
			ListenAddr: "127.0.0.1:6060",
		},

		Ceilometer: ceilometer.NewConfig(),
	}
	return config
}

func loadConfig(configFile string) *Config {
	if configFile == "" {
		fmt.Println("Missing required option `-config /path/to/config.yml`")
		os.Exit(1)
	}

	configContent, err := ioutil.ReadFile(configFile)
	if err != nil {
		fmt.Println("Error reading config file:", err)
		os.Exit(1)
	}

	config := NewConfig()
	err = yaml.Unmarshal(configContent, config)
	if err != nil {
		fmt.Println("Error reading config file:", err)
		os.Exit(1)
	}

	// substitute "{host}" in "Prefix"
	hostname, _ := os.Hostname()
	hostname = strings.Replace(hostname, ".", "-", -1)
	config.Graphite.Prefix = strings.Replace(config.Graphite.Prefix, "{host}", hostname, -1)
	config.Stats.Prefix = strings.Replace(config.Stats.Prefix, "{host}", hostname, -1)

	return config
}
