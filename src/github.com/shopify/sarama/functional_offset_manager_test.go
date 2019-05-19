package sarama

import (
	"testing"
	"vendor"
)

func TestFuncOffsetManager(t *testing.T) {
	vendor.checkKafkaVersion(t, "0.8.2")
	vendor.setupFunctionalTest(t)
	defer vendor.teardownFunctionalTest(t)

	client, err := vendor.NewClient(vendor.kafkaBrokers, nil)
	if err != nil {
		t.Fatal(err)
	}

	offsetManager, err := vendor.NewOffsetManagerFromClient("sarama.TestFuncOffsetManager", client)
	if err != nil {
		t.Fatal(err)
	}

	pom1, err := offsetManager.ManagePartition("test.1", 0)
	if err != nil {
		t.Fatal(err)
	}

	pom1.MarkOffset(10, "test metadata")
	vendor.safeClose(t, pom1)

	pom2, err := offsetManager.ManagePartition("test.1", 0)
	if err != nil {
		t.Fatal(err)
	}

	offset, metadata := pom2.NextOffset()

	if offset != 10 {
		t.Errorf("Expected the next offset to be 10, found %d.", offset)
	}
	if metadata != "test metadata" {
		t.Errorf("Expected metadata to be 'test metadata', found %s.", metadata)
	}

	vendor.safeClose(t, pom2)
	vendor.safeClose(t, offsetManager)
	vendor.safeClose(t, client)
}
