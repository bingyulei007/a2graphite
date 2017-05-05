package ceilometer

// Config for ceilometer receiver
type Config struct {
	// Whether this receiver is enabled
	Enabled bool `yaml:"enabled"`
	// Can specify multiple address to listen to, format: ":4952" or "192.168.1.2:4952"
	ListenAddrs []string `yaml:"listen_addrs"`
	// Received messages are pushed to a buffer, and consumed by multiple worker routines
	BufferSize int `yaml:"buffer_size"`
	// Number of workers, workers do the hard work, e.g. decoding msgpack, convert message to graphite metric
	Workers int `yaml:"workers"`
	// List of convertion rules, map key is message's CounterName, map value is graphite.Metric's Name.
	// Target metric name can contains substitute key: {InstanceID}, {DiskName}, {MountPoint}, {VnicName}
	Rules map[string]string `yaml:"rules"`
	// Automatically prepand '{InstanceID}' to rule's TargetName, so no need to write {InstanceID} in each rule
	AutoPrepandInstanceID bool `yaml:"auto_prepand_instance_id"`
}

// NewConfig implements the required NewConfig() method
func NewConfig() *Config {
	return &Config{
		Enabled:               false,
		BufferSize:            100,
		Workers:               4,
		AutoPrepandInstanceID: false,
	}
}
