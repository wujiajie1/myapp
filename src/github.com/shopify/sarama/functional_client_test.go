package sarama

import (
	"fmt"
	"testing"
	"time"
	"vendor"
)

func TestFuncConnectionFailure(t *testing.T) {
	vendor.setupFunctionalTest(t)
	defer vendor.teardownFunctionalTest(t)

	vendor.Proxies["kafka1"].Enabled = false
	vendor.SaveProxy(t, "kafka1")

	config := vendor.NewConfig()
	config.Metadata.Retry.Max = 1

	_, err := vendor.NewClient([]string{vendor.kafkaBrokers[0]}, config)
	if err != vendor.ErrOutOfBrokers {
		t.Fatal("Expected returned error to be ErrOutOfBrokers, but was: ", err)
	}
}

func TestFuncClientMetadata(t *testing.T) {
	vendor.setupFunctionalTest(t)
	defer vendor.teardownFunctionalTest(t)

	config := vendor.NewConfig()
	config.Metadata.Retry.Max = 1
	config.Metadata.Retry.Backoff = 10 * time.Millisecond
	client, err := vendor.NewClient(vendor.kafkaBrokers, config)
	if err != nil {
		t.Fatal(err)
	}

	if err := client.RefreshMetadata("unknown_topic"); err != vendor.ErrUnknownTopicOrPartition {
		t.Error("Expected ErrUnknownTopicOrPartition, got", err)
	}

	if _, err := client.Leader("unknown_topic", 0); err != vendor.ErrUnknownTopicOrPartition {
		t.Error("Expected ErrUnknownTopicOrPartition, got", err)
	}

	if _, err := client.Replicas("invalid/topic", 0); err != vendor.ErrUnknownTopicOrPartition {
		t.Error("Expected ErrUnknownTopicOrPartition, got", err)
	}

	partitions, err := client.Partitions("test.4")
	if err != nil {
		t.Error(err)
	}
	if len(partitions) != 4 {
		t.Errorf("Expected test.4 topic to have 4 partitions, found %v", partitions)
	}

	partitions, err = client.Partitions("test.1")
	if err != nil {
		t.Error(err)
	}
	if len(partitions) != 1 {
		t.Errorf("Expected test.1 topic to have 1 partitions, found %v", partitions)
	}

	vendor.safeClose(t, client)
}

func TestFuncClientCoordinator(t *testing.T) {
	vendor.checkKafkaVersion(t, "0.8.2")
	vendor.setupFunctionalTest(t)
	defer vendor.teardownFunctionalTest(t)

	client, err := vendor.NewClient(vendor.kafkaBrokers, nil)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 10; i++ {
		broker, err := client.Coordinator(fmt.Sprintf("another_new_consumer_group_%d", i))
		if err != nil {
			t.Fatal(err)
		}

		if connected, err := broker.Connected(); !connected || err != nil {
			t.Errorf("Expected to coordinator %s broker to be properly connected.", broker.Addr())
		}
	}

	vendor.safeClose(t, client)
}
