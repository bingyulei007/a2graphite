package main

import (
	"github.com/openmetric/graphite-client"
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

// All receiver should also provide the following package level methods:

// NewConfig should return point to a config object, the caller will unmarshal config file into this config,
// and pass to NewReceiver().
//type NewConfig func() *Config

// NewReceiver is used to create a receiver object, config is same type of what NewConfig returned.
//type NewReceiver func(config *Config) (*Receiver, error)
