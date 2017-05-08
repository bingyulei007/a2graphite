package ceilometer

// Ceilometer Receiver receives messages published by ceilometer udp publisher.
// https://github.com/openstack/ceilometer/blob/master/ceilometer/publisher/udp.py

import (
	"errors"
	"github.com/op/go-logging"
	"github.com/openmetric/a2graphite/stats"
	"github.com/openmetric/graphite-client"
	"github.com/ugorji/go/codec"
	"net"
	"strings"
	"time"
)

var log = logging.MustGetLogger("ceilometer")

// ceilometerResource represents the received msgpack contents,
// only interested fields are listed.
type ceilometerResource struct {
	CounterName      string      `codec:"counter_name"`
	CounterVolume    interface{} `codec:"counter_volume"`
	CounterUnit      string      `codec:"counter_unit"`
	CounterType      string      `codec:"counter_type"`
	ResourceID       string      `codec:"resource_id"`
	Timestamp        string      `codec:"timestamp"`
	ResourceMetadata struct {
		InstanceID string `codec:"instance_id"`
		DiskName   string `codec:"disk_name"`
		MountPoint string `codec:"mount_point"`
		VnicName   string `codec:"vnic_name"`
	} `codec:"resource_metadata"`
}

// UnixTimestamp returns the Unix timestamp of the message
func (r *ceilometerResource) UnixTimestamp() int64 {
	if t, err := time.Parse("2006-01-02T15:04:05Z", r.Timestamp); err == nil {
		return t.Unix()
	}
	return 0
}

// rule represents convention rule from ceilometerResource to graphite.Metric.
type rule struct {
	TargetName string // the converted graphite Metric's Name

	subInstanceID bool // whether has to substitute InstanceID in Metric's Name
	subDiskName   bool // whether has to substitute DiskName in Metric's Name
	subMountPoint bool // whether has to substitute MountPoint in Metric's Name
	subVnicName   bool // whether has to substitute VnicName in Metric's Name
}

// newRule creates a rule object
func newRule(counterName string, targetName string, autoPrepandInstanceID bool) *rule {
	if autoPrepandInstanceID {
		targetName = "{InstanceID}." + targetName
	}

	r := &rule{
		TargetName: targetName,

		subInstanceID: strings.Contains(targetName, "{InstanceID}"),
		subDiskName:   strings.Contains(targetName, "{DiskName}"),
		subMountPoint: strings.Contains(targetName, "{MountPoint}"),
		subVnicName:   strings.Contains(targetName, "{VnicName}"),
	}

	// check if there are unrecognized substitution

	return r
}

func cleanMountPoint(mountPoint string) string {
	// MountPoint in ceilometer message may contain white space, newline and other illegal characters,
	// trim all these characters in both end, and replace "/"s in middle with "-"
	mountPoint = strings.Trim(mountPoint, "/:\\ \t\r\n")
	if mountPoint == "" {
		return "root"
	}
	mountPoint = strings.Replace(mountPoint, "/", "-", -1)
	return mountPoint
}

// convert ceilometerResource to graphite.Metric
func (r *rule) convert(resource *ceilometerResource) (*graphite.Metric, error) {
	metricName := r.TargetName
	if r.subInstanceID {
		metricName = strings.Replace(metricName, "{InstanceID}", resource.ResourceMetadata.InstanceID, -1)
	}
	if r.subDiskName {
		metricName = strings.Replace(metricName, "{DiskName}", resource.ResourceMetadata.DiskName, -1)
	}
	if r.subMountPoint {
		metricName = strings.Replace(metricName, "{MountPoint}", cleanMountPoint(resource.ResourceMetadata.MountPoint), -1)
	}
	if r.subVnicName {
		metricName = strings.Replace(metricName, "{VnicName}", resource.ResourceMetadata.VnicName, -1)
	}

	metric := &graphite.Metric{
		Name:      metricName,
		Value:     resource.CounterVolume,
		Timestamp: resource.UnixTimestamp(),
	}

	return metric, nil
}

type ceiloStats struct {
	UDPReceived            stats.Counter `stats:"celiometer.udp.Received"`
	UDPReceiveError        stats.Counter `stats:"ceilometer.udp.ReceiveError"`
	UDPBufferSize          stats.Gauge   `stats:"ceilometer.udp.BufferSize"`
	UDPDropped             stats.Counter `stats:"ceilometer.udp.dropped"`
	MsgpackReceived        stats.Counter `stats:"ceilometer.msgpack.Received"`
	MsgpackDecodeOK        stats.Counter `stats:"ceilometer.msgpack.DecodeOK"`
	MsgpackDecodeError     stats.Counter `stats:"ceilometer.msgpack.DecodeError"`
	MetricConvertNoRule    stats.Counter `stats:"ceilometer.convert.NoRule"`
	MetricConvertOK        stats.Counter `stats:"ceilometer.convert.OK"`
	MetricConvertError     stats.Counter `stats:"ceilometer.convert.Error"`
	MetricConvertDiscarded stats.Counter `stats:"ceilometer.convert.Discarded"`
}

// Receiver struct
type Receiver struct {
	config       *Config
	udpMsgBuffer chan []byte
	emitChan     chan *graphite.Metric
	rules        map[string]*rule
	stats        ceiloStats
}

// NewReceiver implements the required NewReceiver() method
func NewReceiver(config interface{}) (*Receiver, error) {
	conf, ok := config.(*Config)
	if !ok {
		return nil, errors.New("invalid type of config parameter")
	}

	if conf.ListenAddrs == nil || len(conf.ListenAddrs) == 0 {
		return nil, errors.New("no `ListenAddrs` provided")
	}

	if conf.Rules == nil || len(conf.Rules) == 0 {
		return nil, errors.New("no `Rules` provided")
	}

	receiver := &Receiver{
		config:       conf,
		udpMsgBuffer: make(chan []byte, conf.BufferSize),
		rules:        make(map[string]*rule),
	}

	for counterName, targetName := range conf.Rules {
		receiver.rules[counterName] = newRule(counterName, targetName, conf.AutoPrepandInstanceID)
		log.Infof("Registered convention rule: %s -> %s\n", counterName, targetName)
	}

	return receiver, nil
}

// GetName implements Receiver's Start method
func (receiver *Receiver) GetName() string {
	return "Ceilometer Receiver"
}

// Start implements Receiver's Start method
func (receiver *Receiver) Start(emitChan chan *graphite.Metric) {
	receiver.emitChan = emitChan

	// NOTE start worker before listener, to prevent buffer pool filled up before workers started
	for i := 0; i < receiver.config.Workers; i++ {
		log.Info("Start Ceilometer Worker:", i)
		go receiver.worker()
	}

	for _, listenAddr := range receiver.config.ListenAddrs {
		go receiver.listener(listenAddr)
	}
}

// Stop implements Receiver's Stop method
func (receiver *Receiver) Stop() {
	// TODO
}

// Healthy implements Receiver's Healthy method
func (receiver *Receiver) Healthy() bool {
	// TODO
	return false
}

// Stats implements Receiver's Stats method
func (receiver *Receiver) Stats() []*graphite.Metric {
	receiver.stats.UDPBufferSize.Set(int64(len(receiver.udpMsgBuffer)))
	return stats.ToGraphiteMetrics(receiver.stats)
}

func (receiver *Receiver) listener(listenAddr string) {
	addr, err := net.ResolveUDPAddr("udp", listenAddr)
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	for {
		// TODO most network interfaces' MTU is less equal to 1500,
		// however it's still possible to receiver big packets.
		// TODO use object pool to improve performance
		raw := make([]byte, 1500)
		if _, _, err := conn.ReadFromUDP(raw); err == nil {
			receiver.stats.UDPReceived.Inc()
			select {
			case receiver.udpMsgBuffer <- raw:
			default:
				receiver.stats.UDPDropped.Inc()
			}
		} else {
			receiver.stats.UDPReceiveError.Inc()
			log.Error("error read from udp", err)
		}
	}
}

func (receiver *Receiver) worker() {
	// NOTE Can `msgpackHandle` be reused safely? Documentation not found.
	msgpackHandle := new(codec.MsgpackHandle)
	for {
		raw := <-receiver.udpMsgBuffer
		receiver.stats.MsgpackReceived.Inc()

		// TODO use object pool to improve performance
		resource := &ceilometerResource{}

		if err := codec.NewDecoderBytes(raw, msgpackHandle).Decode(resource); err != nil {
			receiver.stats.MsgpackDecodeError.Inc()
			log.Error("error unmarshalling msgpack", err)
		} else {
			receiver.stats.MsgpackDecodeOK.Inc()
			metric := receiver.convert(resource)
			if metric != nil {
				receiver.emitChan <- metric
			}
		}
	}
}

func (receiver *Receiver) convert(resource *ceilometerResource) *graphite.Metric {
	if r, ok := receiver.rules[resource.CounterName]; ok {
		if r.TargetName == "" {
			receiver.stats.MetricConvertDiscarded.Inc()
			return nil
		}

		metric, err := r.convert(resource)
		if err == nil {
			receiver.stats.MetricConvertOK.Inc()
			return metric
		}

		receiver.stats.MetricConvertError.Inc()
		log.Error("convert error:", err)
		return nil
	}

	receiver.stats.MetricConvertNoRule.Inc()
	log.Debug("no convert rule found for:", resource.CounterName)
	return nil
}
