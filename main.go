package main

import (
	"github.com/openmetric/graphite-client"
)

// Receiver is interface for all kind of receivers
type Receiver interface {
	// Start the receiver, the receiver should starting listening and receiving metrics,
	// converted graphite metric should be sent back using the provided chan.
	// Start() should not block the caller, only errors preventing the receiver to listen
	// can be returned. If any error returned, the receiver release any resources,
	// e.g. sockets. Stop() won't be called if any error returned.
	Start(chan *graphite.Metric) error

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

// NewReceiver is used to create a receiver object, all receiver packages should provide this method.
type NewReceiver func(config interface{}) *Receiver

// NewConfig should return a config object, the caller will unmarshal config file into this config,
// and pass to NewReceiver().
// All receiver packages should provide this method.
type NewConfig func() interface{}

func main() {

}
