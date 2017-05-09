package main

import (
	"flag"
	"github.com/op/go-logging"
	"github.com/openmetric/a2graphite/ceilometer"
	"github.com/openmetric/graphite-client"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"
)

var log = logging.MustGetLogger("main")

func main() {
	configFile := flag.String("config", "", "Path to the `config file`.")
	flag.Parse()

	config := loadConfig(*configFile)

	// setup log facility
	setupLogging(config.Log)

	var receivers []Receiver
	metrics := make(chan *graphite.Metric, 10000)

	// start graphite client for receiving metrics
	client, err := graphite.NewTCPClient(
		config.Graphite.Host,
		config.Graphite.Port,
		config.Graphite.Prefix,
		config.Graphite.ReconnectDelay,
	)
	if err != nil {
		log.Fatal("Failed to create graphite client:", err)
	}
	go client.SendChan(metrics)

	var statsClient *graphite.Client
	if config.Stats.Enabled {
		// start graphite client for stats
		statsClient, err = graphite.NewTCPClient(
			config.Stats.Host,
			config.Stats.Port,
			config.Stats.Prefix,
			config.Stats.ReconnectDelay,
		)
		if err != nil {
			log.Fatal("Failed to create graphite client for stats:", err)
		}
	}

	// start receivers
	if config.Ceilometer.Enabled {
		if ceilo, err := ceilometer.NewReceiver(config.Ceilometer); err == nil {
			receivers = append(receivers, ceilo)
			ceilo.Start(metrics)
		} else {
			log.Fatal("Failed to initialize ceilometer receiver:", err)
		}
	}

	// check if there are receivers enabled
	if len(receivers) == 0 {
		log.Fatal("You must enable at least one receiver.")
	}

	// enable profiler if configured
	if config.Profiler.Enabled {
		go func() {
			log.Info("Profiler enabled, on:", config.Profiler.ListenAddr)
			log.Info(http.ListenAndServe(config.Profiler.ListenAddr, nil))
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

	healthCheckTicker := time.Tick(1 * time.Second)

	var gcstats debug.GCStats
	for {
		select {
		case <-c:
			log.Info("Got stop signal, stopping ...")
			for _, receiver := range receivers {
				log.Info("Stopping", receiver.GetName(), "...")
				receiver.Stop()
			}
			log.Info("Shuting down graphite client ...")
			client.Shutdown(1 * time.Second)
			log.Info("Shuting down graphite client for stats ...")
			statsClient.Shutdown(1 * time.Second)
			log.Info("Quit.")
			return
		case <-statsTicker:
			for _, receiver := range receivers {
				statsClient.SendMetrics(receiver.Stats())
			}

			debug.ReadGCStats(&gcstats)
			statsClient.SendSimple("gc.NumGC", gcstats.NumGC, 0)
			statsClient.SendSimple("gc.TotalPause", gcstats.PauseTotal, 0)
			if gcstats.NumGC > 0 {
				statsClient.SendSimple("gc.LastPause", gcstats.Pause[0], 0)
			}
		case <-healthCheckTicker:
			for _, receiver := range receivers {
				if !receiver.Healthy() {
					log.Error(receiver.GetName(), "claims to be unhealthy")
				}
			}
		}
	}
}
