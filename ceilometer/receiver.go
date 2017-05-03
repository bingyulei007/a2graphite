package ceilometer

// Ceilometer Receiver receives messages published by ceilometer udp publisher.
// https://github.com/openstack/ceilometer/blob/master/ceilometer/publisher/udp.py

import (
	"github.com/openmetric/graphite-client"
)

// Receiver struct
type Receiver struct {
}

// NewReceiver implements the required NewReceiver() method
func NewReceiver(config Config) *Receiver {

}

// Start implements Receiver's Start method
func (r *Receiver) Start(chan *graphite.Metric) error {
	return nil
}

// Stop implements Receiver's Stop method
func (r *Receiver) Stop() {

}

// Healthy implements Receiver's Healthy method
func (r *Receiver) Healthy() bool {
	return false
}

// Stats implements Receiver's Stats method
func (r *Receiver) Stats() []*graphite.Metric {
	var stats []*graphite.Metric
	return stats
}
