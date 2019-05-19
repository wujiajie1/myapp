package metrics

import (
	"net"
	"time"
	"vendor"
)

func ExampleGraphite() {
	addr, _ := net.ResolveTCPAddr("net", ":2003")
	go vendor.Graphite(vendor.DefaultRegistry, 1*time.Second, "some.prefix", addr)
}

func ExampleGraphiteWithConfig() {
	addr, _ := net.ResolveTCPAddr("net", ":2003")
	go vendor.GraphiteWithConfig(vendor.GraphiteConfig{
		Addr:          addr,
		Registry:      vendor.DefaultRegistry,
		FlushInterval: 1 * time.Second,
		DurationUnit:  time.Millisecond,
		Percentiles:   []float64{0.5, 0.75, 0.99, 0.999},
	})
}
