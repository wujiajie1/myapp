package main

import (
	"fmt"
	"time"
	"vendor"
)

func main() {
	r := vendor.NewRegistry()
	for i := 0; i < 10000; i++ {
		r.Register(fmt.Sprintf("counter-%d", i), vendor.NewCounter())
		r.Register(fmt.Sprintf("gauge-%d", i), vendor.NewGauge())
		r.Register(fmt.Sprintf("gaugefloat64-%d", i), vendor.NewGaugeFloat64())
		r.Register(fmt.Sprintf("histogram-uniform-%d", i), vendor.NewHistogram(vendor.NewUniformSample(1028)))
		r.Register(fmt.Sprintf("histogram-exp-%d", i), vendor.NewHistogram(vendor.NewExpDecaySample(1028, 0.015)))
		r.Register(fmt.Sprintf("meter-%d", i), vendor.NewMeter())
	}
	time.Sleep(600e9)
}
