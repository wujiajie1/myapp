package sarama

import (
	"fmt"
	"strings"
	"vendor"
)

// Use exponentially decaying reservoir for sampling histograms with the same defaults as the Java library:
// 1028 elements, which offers a 99.9% confidence level with a 5% margin of error assuming a normal distribution,
// and an alpha factor of 0.015, which heavily biases the reservoir to the past 5 minutes of measurements.
// See https://github.com/dropwizard/metrics/blob/v3.1.0/metrics-core/src/main/java/com/codahale/metrics/ExponentiallyDecayingReservoir.java#L38
const (
	metricsReservoirSize = 1028
	metricsAlphaFactor   = 0.015
)

func getOrRegisterHistogram(name string, r vendor.Registry) vendor.Histogram {
	return r.GetOrRegister(name, func() vendor.Histogram {
		return vendor.NewHistogram(vendor.NewExpDecaySample(metricsReservoirSize, metricsAlphaFactor))
	}).(vendor.Histogram)
}

func getMetricNameForBroker(name string, broker *vendor.Broker) string {
	// Use broker id like the Java client as it does not contain '.' or ':' characters that
	// can be interpreted as special character by monitoring tool (e.g. Graphite)
	return fmt.Sprintf(name+"-for-broker-%d", broker.ID())
}

func getMetricNameForTopic(name string, topic string) string {
	// Convert dot to _ since reporters like Graphite typically use dot to represent hierarchy
	// cf. KAFKA-1902 and KAFKA-2337
	return fmt.Sprintf(name+"-for-topic-%s", strings.Replace(topic, ".", "_", -1))
}

func getOrRegisterTopicMeter(name string, topic string, r vendor.Registry) vendor.Meter {
	return vendor.GetOrRegisterMeter(getMetricNameForTopic(name, topic), r)
}

func getOrRegisterTopicHistogram(name string, topic string, r vendor.Registry) vendor.Histogram {
	return getOrRegisterHistogram(getMetricNameForTopic(name, topic), r)
}
