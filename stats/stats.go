package stats

import (
	"github.com/bingyulei007/graphite-client"
	"reflect"
	"sync/atomic"
	"time"
)

// Counter type for counter type metrics
type Counter struct {
	value uint64
}

// Add n to the counter
func (c *Counter) Add(n uint64) {
	atomic.AddUint64(&c.value, n)
}

// Inc the counter by 1
func (c *Counter) Inc() {
	atomic.AddUint64(&c.value, 1)
}

// Load counter's value
func (c *Counter) Load() uint64 {
	return atomic.LoadUint64(&c.value)
}

// Gauge type for gauge type metrics
type Gauge struct {
	value int64
}

// Set value of gauge
func (g *Gauge) Set(n int64) {
	atomic.StoreInt64(&g.value, n)
}

// Load gauge's value
func (g *Gauge) Load() int64 {
	return atomic.LoadInt64(&g.value)
}

// ToGraphiteMetrics converts a struct of Counter type fields into graphite.Metric
// Example:
//
//     type Stats struct {
//         A Counter `stats:"a2graphite.stat.a"`
//         B Counter `stats:"other.b"`
//     }
//
// will yield two metrics, metric name is specified via `stats` tag.
func ToGraphiteMetrics(s interface{}) []*graphite.Metric {
	val := reflect.ValueOf(s)
	typ := val.Type()
	numFields := val.NumField()
	var results []*graphite.Metric
	timestamp := time.Now().Unix()

	for i := 0; i < numFields; i++ {
		switch value := val.Field(i).Interface().(type) {
		case Counter:
			results = append(results, &graphite.Metric{
				Name:      typ.Field(i).Tag.Get("stats"),
				Value:     value.Load(),
				Timestamp: timestamp,
			})
		case Gauge:
			results = append(results, &graphite.Metric{
				Name:      typ.Field(i).Tag.Get("stats"),
				Value:     value.Load(),
				Timestamp: timestamp,
			})
		}
	}
	return results
}
