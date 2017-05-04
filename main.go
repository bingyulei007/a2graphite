package main

import (
	"flag"
	"github.com/openmetric/a2graphite/ceilometer"
	"github.com/openmetric/graphite-client"
	"github.com/spf13/viper"
	"log"
	"os"
	"time"
)

// Receiver is interface for all kind of receivers
type Receiver interface {
	// Start the receiver, the receiver should starting listening and receiving metrics,
	// converted graphite metric should be sent back using the provided chan.
	// Start() should not block the caller.
	Start(chan *graphite.Metric)

	// Stop the receiver, the receiver should stop listening and receiving metrics.
	// Stop() could block the caller, but the caller does not guarantee to wait for Stop()
	// to return, if the caller waited for a reasonable timeout, receiver routines may be
	// killed forcely.
	Stop()

	// Check if the receiver is still working (for health check).
	Healthy() bool

	// Stats returns the internal stats. This method is called periodically.
	// The caller will prefix metrics' names with "${user-configured-prefix}.${receiver}",
	// so the receiver does not need to prefix these stuffs.
	Stats() []*graphite.Metric
}

// NewConfig should return point to a config object, the caller will unmarshal config file into this config,
// and pass to NewReceiver().
// All receiver packages should provide this method.
type NewConfig func() interface{}

// NewReceiver is used to create a receiver object, all receiver packages should provide this method.
// config is same type of what NewConfig returned.
type NewReceiver func(config interface{}) (*Receiver, error)

func main() {
	configFile := flag.String("config", "", "Path to the config file.")
	flag.Parse()

	if *configFile == "" {
		log.Fatal("Missing required option `-config /path/to/config.yml`")
	}

	if _, err := os.Stat(*configFile); os.IsNotExist(err) {
		log.Fatal("Config file does not exist:", *configFile)
	}

	viper.SetConfigFile(*configFile)
	viper.SetConfigType("yaml")
	viper.ReadInConfig()

	metricPool := make(chan *graphite.Metric, 1000)
	graphiteClient, err := graphite.NewTCPClient("127.0.0.1", 2003, "", 1*time.Microsecond)
	if err != nil {
		log.Fatalln(err)
	}

	ceilometerConfig := ceilometer.NewConfig()
	viper.UnmarshalKey("ceilometer", ceilometerConfig)

	ceilometerReceiver, err := ceilometer.NewReceiver(ceilometerConfig)
	if err != nil {
		log.Fatalln("failed to create ceilometer receiver:", err)
	}
	ceilometerReceiver.Start(metricPool)

	for {
		metric := <-metricPool
		graphiteClient.SendMetric(metric)
	}
}
