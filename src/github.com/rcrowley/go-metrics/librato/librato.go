package librato

import (
	"fmt"
	"log"
	"math"
	"regexp"
	"time"
	"vendor"
)

// a regexp for extracting the unit from time.Duration.String
var unitRegexp = regexp.MustCompile("[^\\d]+$")

// a helper that turns a time.Duration into librato display attributes for timer metrics
func translateTimerAttributes(d time.Duration) (attrs map[string]interface{}) {
	attrs = make(map[string]interface{})
	attrs[vendor.DisplayTransform] = fmt.Sprintf("x/%d", int64(d))
	attrs[vendor.DisplayUnitsShort] = string(unitRegexp.Find([]byte(d.String())))
	return
}

type Reporter struct {
	Email, Token    string
	Namespace       string
	Source          string
	Interval        time.Duration
	Registry        vendor.Registry
	Percentiles     []float64              // percentiles to report on histogram metrics
	TimerAttributes map[string]interface{} // units in which timers will be displayed
	intervalSec     int64
}

func NewReporter(r vendor.Registry, d time.Duration, e string, t string, s string, p []float64, u time.Duration) *Reporter {
	return &Reporter{e, t, "", s, d, r, p, translateTimerAttributes(u), int64(d / time.Second)}
}

func Librato(r vendor.Registry, d time.Duration, e string, t string, s string, p []float64, u time.Duration) {
	NewReporter(r, d, e, t, s, p, u).Run()
}

func (self *Reporter) Run() {
	log.Printf("WARNING: This client has been DEPRECATED! It has been moved to https://github.com/mihasya/go-metrics-librato and will be removed from rcrowley/go-metrics on August 5th 2015")
	ticker := time.Tick(self.Interval)
	metricsApi := &vendor.LibratoClient{self.Email, self.Token}
	for now := range ticker {
		var metrics vendor.Batch
		var err error
		if metrics, err = self.BuildRequest(now, self.Registry); err != nil {
			log.Printf("ERROR constructing librato request body %s", err)
			continue
		}
		if err := metricsApi.PostMetrics(metrics); err != nil {
			log.Printf("ERROR sending metrics to librato %s", err)
			continue
		}
	}
}

// calculate sum of squares from data provided by metrics.Histogram
// see http://en.wikipedia.org/wiki/Standard_deviation#Rapid_calculation_methods
func sumSquares(s vendor.Sample) float64 {
	count := float64(s.Count())
	sumSquared := math.Pow(count*s.Mean(), 2)
	sumSquares := math.Pow(count*s.StdDev(), 2) + sumSquared/count
	if math.IsNaN(sumSquares) {
		return 0.0
	}
	return sumSquares
}
func sumSquaresTimer(t vendor.Timer) float64 {
	count := float64(t.Count())
	sumSquared := math.Pow(count*t.Mean(), 2)
	sumSquares := math.Pow(count*t.StdDev(), 2) + sumSquared/count
	if math.IsNaN(sumSquares) {
		return 0.0
	}
	return sumSquares
}

func (self *Reporter) BuildRequest(now time.Time, r vendor.Registry) (snapshot vendor.Batch, err error) {
	snapshot = vendor.Batch{
		// coerce timestamps to a stepping fn so that they line up in Librato graphs
		MeasureTime: (now.Unix() / self.intervalSec) * self.intervalSec,
		Source:      self.Source,
	}
	snapshot.Gauges = make([]vendor.Measurement, 0)
	snapshot.Counters = make([]vendor.Measurement, 0)
	histogramGaugeCount := 1 + len(self.Percentiles)
	r.Each(func(name string, metric interface{}) {
		if self.Namespace != "" {
			name = fmt.Sprintf("%s.%s", self.Namespace, name)
		}
		measurement := vendor.Measurement{}
		measurement[vendor.Period] = self.Interval.Seconds()
		switch m := metric.(type) {
		case vendor.Counter:
			if m.Count() > 0 {
				measurement[vendor.Name] = fmt.Sprintf("%s.%s", name, "count")
				measurement[vendor.Value] = float64(m.Count())
				measurement[vendor.Attributes] = map[string]interface{}{
					vendor.DisplayUnitsLong:  vendor.Operations,
					vendor.DisplayUnitsShort: vendor.OperationsShort,
					vendor.DisplayMin:        "0",
				}
				snapshot.Counters = append(snapshot.Counters, measurement)
			}
		case vendor.Gauge:
			measurement[vendor.Name] = name
			measurement[vendor.Value] = float64(m.Value())
			snapshot.Gauges = append(snapshot.Gauges, measurement)
		case vendor.GaugeFloat64:
			measurement[vendor.Name] = name
			measurement[vendor.Value] = float64(m.Value())
			snapshot.Gauges = append(snapshot.Gauges, measurement)
		case vendor.Histogram:
			if m.Count() > 0 {
				gauges := make([]vendor.Measurement, histogramGaugeCount, histogramGaugeCount)
				s := m.Sample()
				measurement[vendor.Name] = fmt.Sprintf("%s.%s", name, "hist")
				measurement[vendor.Count] = uint64(s.Count())
				measurement[vendor.Max] = float64(s.Max())
				measurement[vendor.Min] = float64(s.Min())
				measurement[vendor.Sum] = float64(s.Sum())
				measurement[vendor.SumSquares] = sumSquares(s)
				gauges[0] = measurement
				for i, p := range self.Percentiles {
					gauges[i+1] = vendor.Measurement{
						vendor.Name:   fmt.Sprintf("%s.%.2f", measurement[vendor.Name], p),
						vendor.Value:  s.Percentile(p),
						vendor.Period: measurement[vendor.Period],
					}
				}
				snapshot.Gauges = append(snapshot.Gauges, gauges...)
			}
		case vendor.Meter:
			measurement[vendor.Name] = name
			measurement[vendor.Value] = float64(m.Count())
			snapshot.Counters = append(snapshot.Counters, measurement)
			snapshot.Gauges = append(snapshot.Gauges,
				vendor.Measurement{
					vendor.Name:   fmt.Sprintf("%s.%s", name, "1min"),
					vendor.Value:  m.Rate1(),
					vendor.Period: int64(self.Interval.Seconds()),
					vendor.Attributes: map[string]interface{}{
						vendor.DisplayUnitsLong:  vendor.Operations,
						vendor.DisplayUnitsShort: vendor.OperationsShort,
						vendor.DisplayMin:        "0",
					},
				},
				vendor.Measurement{
					vendor.Name:   fmt.Sprintf("%s.%s", name, "5min"),
					vendor.Value:  m.Rate5(),
					vendor.Period: int64(self.Interval.Seconds()),
					vendor.Attributes: map[string]interface{}{
						vendor.DisplayUnitsLong:  vendor.Operations,
						vendor.DisplayUnitsShort: vendor.OperationsShort,
						vendor.DisplayMin:        "0",
					},
				},
				vendor.Measurement{
					vendor.Name:   fmt.Sprintf("%s.%s", name, "15min"),
					vendor.Value:  m.Rate15(),
					vendor.Period: int64(self.Interval.Seconds()),
					vendor.Attributes: map[string]interface{}{
						vendor.DisplayUnitsLong:  vendor.Operations,
						vendor.DisplayUnitsShort: vendor.OperationsShort,
						vendor.DisplayMin:        "0",
					},
				},
			)
		case vendor.Timer:
			measurement[vendor.Name] = name
			measurement[vendor.Value] = float64(m.Count())
			snapshot.Counters = append(snapshot.Counters, measurement)
			if m.Count() > 0 {
				libratoName := fmt.Sprintf("%s.%s", name, "timer.mean")
				gauges := make([]vendor.Measurement, histogramGaugeCount, histogramGaugeCount)
				gauges[0] = vendor.Measurement{
					vendor.Name:       libratoName,
					vendor.Count:      uint64(m.Count()),
					vendor.Sum:        m.Mean() * float64(m.Count()),
					vendor.Max:        float64(m.Max()),
					vendor.Min:        float64(m.Min()),
					vendor.SumSquares: sumSquaresTimer(m),
					vendor.Period:     int64(self.Interval.Seconds()),
					vendor.Attributes: self.TimerAttributes,
				}
				for i, p := range self.Percentiles {
					gauges[i+1] = vendor.Measurement{
						vendor.Name:       fmt.Sprintf("%s.timer.%2.0f", name, p*100),
						vendor.Value:      m.Percentile(p),
						vendor.Period:     int64(self.Interval.Seconds()),
						vendor.Attributes: self.TimerAttributes,
					}
				}
				snapshot.Gauges = append(snapshot.Gauges, gauges...)
				snapshot.Gauges = append(snapshot.Gauges,
					vendor.Measurement{
						vendor.Name:   fmt.Sprintf("%s.%s", name, "rate.1min"),
						vendor.Value:  m.Rate1(),
						vendor.Period: int64(self.Interval.Seconds()),
						vendor.Attributes: map[string]interface{}{
							vendor.DisplayUnitsLong:  vendor.Operations,
							vendor.DisplayUnitsShort: vendor.OperationsShort,
							vendor.DisplayMin:        "0",
						},
					},
					vendor.Measurement{
						vendor.Name:   fmt.Sprintf("%s.%s", name, "rate.5min"),
						vendor.Value:  m.Rate5(),
						vendor.Period: int64(self.Interval.Seconds()),
						vendor.Attributes: map[string]interface{}{
							vendor.DisplayUnitsLong:  vendor.Operations,
							vendor.DisplayUnitsShort: vendor.OperationsShort,
							vendor.DisplayMin:        "0",
						},
					},
					vendor.Measurement{
						vendor.Name:   fmt.Sprintf("%s.%s", name, "rate.15min"),
						vendor.Value:  m.Rate15(),
						vendor.Period: int64(self.Interval.Seconds()),
						vendor.Attributes: map[string]interface{}{
							vendor.DisplayUnitsLong:  vendor.Operations,
							vendor.DisplayUnitsShort: vendor.OperationsShort,
							vendor.DisplayMin:        "0",
						},
					},
				)
			}
		}
	})
	return
}
