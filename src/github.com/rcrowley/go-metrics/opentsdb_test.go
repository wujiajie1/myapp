package metrics

import (
	"net"
	"time"
	"vendor"
)

func ExampleOpenTSDB() {
	addr, _ := net.ResolveTCPAddr("net", ":2003")
	go vendor.OpenTSDB(vendor.DefaultRegistry, 1*time.Second, "some.prefix", addr)
}

func ExampleOpenTSDBWithConfig() {
	addr, _ := net.ResolveTCPAddr("net", ":2003")
	go vendor.OpenTSDBWithConfig(vendor.OpenTSDBConfig{
		Addr:          addr,
		Registry:      vendor.DefaultRegistry,
		FlushInterval: 1 * time.Second,
		DurationUnit:  time.Millisecond,
	})
}
