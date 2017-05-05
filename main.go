package main

import (
	"flag"
	"github.com/openmetric/a2graphite/ceilometer"
	"github.com/openmetric/graphite-client"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	configFile := flag.String("config", "", "Path to the `config file`.")
	flag.Parse()

	config := loadConfig(*configFile)

	var receivers []Receiver
	metrics := make(chan *graphite.Metric, 10000)

	// start graphite client for receiving metrics
	client, err := graphite.NewTCPClient(
		config.Graphite.GraphiteHost,
		config.Graphite.GraphitePort,
		config.Graphite.Prefix,
		config.Graphite.ReconnectDelay,
	)
	if err != nil {
		log.Fatalln("Failed to create graphite client:", err)
	}
	go client.SendChan(metrics)

	var statsClient *graphite.Client
	if config.Stats.Enabled {
		// start graphite client for stats
		statsClient, err = graphite.NewTCPClient(
			config.Stats.GraphiteHost,
			config.Stats.GraphitePort,
			config.Stats.Prefix,
			config.Stats.ReconnectDelay,
		)
		if err != nil {
			log.Fatalln("Failed to create graphite client for stats:", err)
		}
	}

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

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	var statsTicker <-chan time.Time
	if config.Stats.Enabled {
		statsTicker = time.Tick(config.Stats.Interval)
	} else {
		statsTicker = make(chan time.Time)
	}

	for {
		select {
		case <-c:
			log.Println("Got stop signal, stopping...")
			for _, receiver := range receivers {
				log.Println("Stopping", receiver.GetName())
				receiver.Stop()
			}
			log.Println("Shuting down graphite client ...")
			client.Shutdown(1 * time.Second)
			log.Println("Shuting down graphite client for stats ...")
			statsClient.Shutdown(1 * time.Second)
			log.Println("Quit.")
			return
		case <-statsTicker:
			for _, receiver := range receivers {
				statsClient.SendMetrics(receiver.Stats())
			}
		}
	}
}
