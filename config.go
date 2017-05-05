package main

import (
	"github.com/openmetric/a2graphite/ceilometer"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"time"
)

type graphiteConfig struct {
	GraphiteHost   string        `yaml:"graphite_host"`
	GraphitePort   int           `yaml:"graphite_port"`
	ReconnectDelay time.Duration `yaml:"reconnect_delay"`
	Prefix         string        `yaml:"prefix"`
	BufferSize     int           `yaml:"buffer_size"`
}

type statsConfig struct {
	GraphiteHost   string        `yaml:"graphite_host"`
	GraphitePort   int           `yaml:"graphite_port"`
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
	Graphite *graphiteConfig `yaml:"graphite"`
	Stats    *statsConfig    `yaml:"stats"`
	Profiler *profilerConfig `yaml:"profiler"`

	// Receiver configs
	Ceilometer *ceilometer.Config `yaml:"ceilometer"`
}

// NewConfig returns a default config instance
func NewConfig() *Config {
	config := &Config{
		Graphite: &graphiteConfig{
			ReconnectDelay: 100 * time.Microsecond,
			BufferSize:     100,
		},
		Stats: &statsConfig{
			ReconnectDelay: 100 * time.Microsecond,
			BufferSize:     100,
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
		log.Fatal("Missing required option `-config /path/to/config.yml`")
	}

	configContent, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatalln("Error reading config file:", err)
	}

	config := NewConfig()
	err = yaml.Unmarshal(configContent, config)
	if err != nil {
		log.Fatalln("Error reading config file:", err)
	}

	return config
}
