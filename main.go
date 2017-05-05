package main

import (
	"flag"
	"github.com/openmetric/a2graphite/ceilometer"
	"github.com/openmetric/graphite-client"
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"
)

func main() {
	configFile := flag.String("config", "", "Path to the `config file`.")
	flag.Parse()

	config := loadConfig(*configFile)

	var receivers []Receiver
	metrics := make(chan *graphite.Metric, 10000)

	// start graphite client
	client, err := graphite.NewTCPClient(
		config.Graphite.GraphiteHost,
		config.Graphite.GraphitePort,
		config.Graphite.Prefix,
		time.Second*1,
	)
	if err != nil {
		log.Fatalln("Failed to create graphite client:", err)
	}
	go client.SendChan(metrics)

	// start receivers
	if config.Ceilometer.Enabled {
		if ceilo, err := ceilometer.NewReceiver(config.Ceilometer); err == nil {
			receivers = append(receivers, ceilo)
			ceilo.Start(metrics)
		} else {
			log.Fatalln("Failed to initialize ceilometer receiver:", err)
		}
	}

	// check if there are receivers enabled
	if len(receivers) == 0 {
		log.Fatalln("You must enable at least one receiver.")
	}

	// enable profiler if configured
	if config.Profiler.Enabled {
		go func() {
			log.Println("Profiled enabled, on:", config.Profiler.ListenAddr)
			log.Println(http.ListenAndServe(config.Profiler.ListenAddr, nil))
		}()
	}

	// TODO pull stats from modules

	// TODO catch SIGTERM and SIGINT, stop gracefully

	for {
		time.Sleep(1)
	}
}
