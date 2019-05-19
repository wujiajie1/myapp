package sarama

import (
	"sync/atomic"
	"testing"
	"time"
	"vendor"
)

func initOffsetManagerWithBackoffFunc(t *testing.T, retention time.Duration,
	backoffFunc func(retries, maxRetries int) time.Duration) (om vendor.OffsetManager,
	testClient vendor.Client, broker, coordinator *vendor.MockBroker) {

	config := vendor.NewConfig()
	config.Metadata.Retry.Max = 1
	if backoffFunc != nil {
		config.Metadata.Retry.BackoffFunc = backoffFunc
	}
	config.Consumer.Offsets.CommitInterval = 1 * time.Millisecond
	config.Version = vendor.V0_9_0_0
	if retention > 0 {
		config.Consumer.Offsets.Retention = retention
	}

	broker = vendor.NewMockBroker(t, 1)
	coordinator = vendor.NewMockBroker(t, 2)

	seedMeta := new(vendor.MetadataResponse)
	seedMeta.AddBroker(coordinator.Addr(), coordinator.BrokerID())
	seedMeta.AddTopicPartition("my_topic", 0, 1, []int32{}, []int32{}, []int32{}, vendor.ErrNoError)
	seedMeta.AddTopicPartition("my_topic", 1, 1, []int32{}, []int32{}, []int32{}, vendor.ErrNoError)
	broker.Returns(seedMeta)

	var err error
	testClient, err = vendor.NewClient([]string{broker.Addr()}, config)
	if err != nil {
		t.Fatal(err)
	}

	broker.Returns(&vendor.ConsumerMetadataResponse{
		CoordinatorID:   coordinator.BrokerID(),
		CoordinatorHost: "127.0.0.1",
		CoordinatorPort: coordinator.Port(),
	})

	om, err = vendor.NewOffsetManagerFromClient("group", testClient)
	if err != nil {
		t.Fatal(err)
	}

	return om, testClient, broker, coordinator
}

func initOffsetManager(t *testing.T, retention time.Duration) (om vendor.OffsetManager,
	testClient vendor.Client, broker, coordinator *vendor.MockBroker) {
	return initOffsetManagerWithBackoffFunc(t, retention, nil)
}

func initPartitionOffsetManager(t *testing.T, om vendor.OffsetManager,
	coordinator *vendor.MockBroker, initialOffset int64, metadata string) vendor.PartitionOffsetManager {

	fetchResponse := new(vendor.OffsetFetchResponse)
	fetchResponse.AddBlock("my_topic", 0, &vendor.OffsetFetchResponseBlock{
		Err:      vendor.ErrNoError,
		Offset:   initialOffset,
		Metadata: metadata,
	})
	coordinator.Returns(fetchResponse)

	pom, err := om.ManagePartition("my_topic", 0)
	if err != nil {
		t.Fatal(err)
	}

	return pom
}

func TestNewOffsetManager(t *testing.T) {
	seedBroker := vendor.NewMockBroker(t, 1)
	seedBroker.Returns(new(vendor.MetadataResponse))
	defer seedBroker.Close()

	testClient, err := vendor.NewClient([]string{seedBroker.Addr()}, nil)
	if err != nil {
		t.Fatal(err)
	}

	om, err := vendor.NewOffsetManagerFromClient("group", testClient)
	if err != nil {
		t.Error(err)
	}
	vendor.safeClose(t, om)
	vendor.safeClose(t, testClient)

	_, err = vendor.NewOffsetManagerFromClient("group", testClient)
	if err != vendor.ErrClosedClient {
		t.Errorf("Error expected for closed client; actual value: %v", err)
	}
}

// Test recovery from ErrNotCoordinatorForConsumer
// on first fetchInitialOffset call
func TestOffsetManagerFetchInitialFail(t *testing.T) {
	om, testClient, broker, coordinator := initOffsetManager(t, 0)

	// Error on first fetchInitialOffset call
	responseBlock := vendor.OffsetFetchResponseBlock{
		Err:      vendor.ErrNotCoordinatorForConsumer,
		Offset:   5,
		Metadata: "test_meta",
	}

	fetchResponse := new(vendor.OffsetFetchResponse)
	fetchResponse.AddBlock("my_topic", 0, &responseBlock)
	coordinator.Returns(fetchResponse)

	// Refresh coordinator
	newCoordinator := vendor.NewMockBroker(t, 3)
	broker.Returns(&vendor.ConsumerMetadataResponse{
		CoordinatorID:   newCoordinator.BrokerID(),
		CoordinatorHost: "127.0.0.1",
		CoordinatorPort: newCoordinator.Port(),
	})

	// Second fetchInitialOffset call is fine
	fetchResponse2 := new(vendor.OffsetFetchResponse)
	responseBlock2 := responseBlock
	responseBlock2.Err = vendor.ErrNoError
	fetchResponse2.AddBlock("my_topic", 0, &responseBlock2)
	newCoordinator.Returns(fetchResponse2)

	pom, err := om.ManagePartition("my_topic", 0)
	if err != nil {
		t.Error(err)
	}

	broker.Close()
	coordinator.Close()
	newCoordinator.Close()
	vendor.safeClose(t, pom)
	vendor.safeClose(t, om)
	vendor.safeClose(t, testClient)
}

// Test fetchInitialOffset retry on ErrOffsetsLoadInProgress
func TestOffsetManagerFetchInitialLoadInProgress(t *testing.T) {
	retryCount := int32(0)
	backoff := func(retries, maxRetries int) time.Duration {
		atomic.AddInt32(&retryCount, 1)
		return 0
	}
	om, testClient, broker, coordinator := initOffsetManagerWithBackoffFunc(t, 0, backoff)

	// Error on first fetchInitialOffset call
	responseBlock := vendor.OffsetFetchResponseBlock{
		Err:      vendor.ErrOffsetsLoadInProgress,
		Offset:   5,
		Metadata: "test_meta",
	}

	fetchResponse := new(vendor.OffsetFetchResponse)
	fetchResponse.AddBlock("my_topic", 0, &responseBlock)
	coordinator.Returns(fetchResponse)

	// Second fetchInitialOffset call is fine
	fetchResponse2 := new(vendor.OffsetFetchResponse)
	responseBlock2 := responseBlock
	responseBlock2.Err = vendor.ErrNoError
	fetchResponse2.AddBlock("my_topic", 0, &responseBlock2)
	coordinator.Returns(fetchResponse2)

	pom, err := om.ManagePartition("my_topic", 0)
	if err != nil {
		t.Error(err)
	}

	broker.Close()
	coordinator.Close()
	vendor.safeClose(t, pom)
	vendor.safeClose(t, om)
	vendor.safeClose(t, testClient)

	if atomic.LoadInt32(&retryCount) == 0 {
		t.Fatal("Expected at least one retry")
	}
}

func TestPartitionOffsetManagerInitialOffset(t *testing.T) {
	om, testClient, broker, coordinator := initOffsetManager(t, 0)
	testClient.Config().Consumer.Offsets.Initial = vendor.OffsetOldest

	// Kafka returns -1 if no offset has been stored for this partition yet.
	pom := initPartitionOffsetManager(t, om, coordinator, -1, "")

	offset, meta := pom.NextOffset()
	if offset != vendor.OffsetOldest {
		t.Errorf("Expected offset 5. Actual: %v", offset)
	}
	if meta != "" {
		t.Errorf("Expected metadata to be empty. Actual: %q", meta)
	}

	vendor.safeClose(t, pom)
	vendor.safeClose(t, om)
	broker.Close()
	coordinator.Close()
	vendor.safeClose(t, testClient)
}

func TestPartitionOffsetManagerNextOffset(t *testing.T) {
	om, testClient, broker, coordinator := initOffsetManager(t, 0)
	pom := initPartitionOffsetManager(t, om, coordinator, 5, "test_meta")

	offset, meta := pom.NextOffset()
	if offset != 5 {
		t.Errorf("Expected offset 5. Actual: %v", offset)
	}
	if meta != "test_meta" {
		t.Errorf("Expected metadata \"test_meta\". Actual: %q", meta)
	}

	vendor.safeClose(t, pom)
	vendor.safeClose(t, om)
	broker.Close()
	coordinator.Close()
	vendor.safeClose(t, testClient)
}

func TestPartitionOffsetManagerResetOffset(t *testing.T) {
	om, testClient, broker, coordinator := initOffsetManager(t, 0)
	pom := initPartitionOffsetManager(t, om, coordinator, 5, "original_meta")

	ocResponse := new(vendor.OffsetCommitResponse)
	ocResponse.AddError("my_topic", 0, vendor.ErrNoError)
	coordinator.Returns(ocResponse)

	expected := int64(1)
	pom.ResetOffset(expected, "modified_meta")
	actual, meta := pom.NextOffset()

	if actual != expected {
		t.Errorf("Expected offset %v. Actual: %v", expected, actual)
	}
	if meta != "modified_meta" {
		t.Errorf("Expected metadata \"modified_meta\". Actual: %q", meta)
	}

	vendor.safeClose(t, pom)
	vendor.safeClose(t, om)
	vendor.safeClose(t, testClient)
	broker.Close()
	coordinator.Close()
}

func TestPartitionOffsetManagerResetOffsetWithRetention(t *testing.T) {
	om, testClient, broker, coordinator := initOffsetManager(t, time.Hour)
	pom := initPartitionOffsetManager(t, om, coordinator, 5, "original_meta")

	ocResponse := new(vendor.OffsetCommitResponse)
	ocResponse.AddError("my_topic", 0, vendor.ErrNoError)
	handler := func(req *vendor.request) (res vendor.encoder) {
		if req.body.version() != 2 {
			t.Errorf("Expected to be using version 2. Actual: %v", req.body.version())
		}
		offsetCommitRequest := req.body.(*vendor.OffsetCommitRequest)
		if offsetCommitRequest.RetentionTime != (60 * 60 * 1000) {
			t.Errorf("Expected an hour retention time. Actual: %v", offsetCommitRequest.RetentionTime)
		}
		return ocResponse
	}
	coordinator.setHandler(handler)

	expected := int64(1)
	pom.ResetOffset(expected, "modified_meta")
	actual, meta := pom.NextOffset()

	if actual != expected {
		t.Errorf("Expected offset %v. Actual: %v", expected, actual)
	}
	if meta != "modified_meta" {
		t.Errorf("Expected metadata \"modified_meta\". Actual: %q", meta)
	}

	vendor.safeClose(t, pom)
	vendor.safeClose(t, om)
	vendor.safeClose(t, testClient)
	broker.Close()
	coordinator.Close()
}

func TestPartitionOffsetManagerMarkOffset(t *testing.T) {
	om, testClient, broker, coordinator := initOffsetManager(t, 0)
	pom := initPartitionOffsetManager(t, om, coordinator, 5, "original_meta")

	ocResponse := new(vendor.OffsetCommitResponse)
	ocResponse.AddError("my_topic", 0, vendor.ErrNoError)
	coordinator.Returns(ocResponse)

	pom.MarkOffset(100, "modified_meta")
	offset, meta := pom.NextOffset()

	if offset != 100 {
		t.Errorf("Expected offset 100. Actual: %v", offset)
	}
	if meta != "modified_meta" {
		t.Errorf("Expected metadata \"modified_meta\". Actual: %q", meta)
	}

	vendor.safeClose(t, pom)
	vendor.safeClose(t, om)
	vendor.safeClose(t, testClient)
	broker.Close()
	coordinator.Close()
}

func TestPartitionOffsetManagerMarkOffsetWithRetention(t *testing.T) {
	om, testClient, broker, coordinator := initOffsetManager(t, time.Hour)
	pom := initPartitionOffsetManager(t, om, coordinator, 5, "original_meta")

	ocResponse := new(vendor.OffsetCommitResponse)
	ocResponse.AddError("my_topic", 0, vendor.ErrNoError)
	handler := func(req *vendor.request) (res vendor.encoder) {
		if req.body.version() != 2 {
			t.Errorf("Expected to be using version 2. Actual: %v", req.body.version())
		}
		offsetCommitRequest := req.body.(*vendor.OffsetCommitRequest)
		if offsetCommitRequest.RetentionTime != (60 * 60 * 1000) {
			t.Errorf("Expected an hour retention time. Actual: %v", offsetCommitRequest.RetentionTime)
		}
		return ocResponse
	}
	coordinator.setHandler(handler)

	pom.MarkOffset(100, "modified_meta")
	offset, meta := pom.NextOffset()

	if offset != 100 {
		t.Errorf("Expected offset 100. Actual: %v", offset)
	}
	if meta != "modified_meta" {
		t.Errorf("Expected metadata \"modified_meta\". Actual: %q", meta)
	}

	vendor.safeClose(t, pom)
	vendor.safeClose(t, om)
	vendor.safeClose(t, testClient)
	broker.Close()
	coordinator.Close()
}

func TestPartitionOffsetManagerCommitErr(t *testing.T) {
	om, testClient, broker, coordinator := initOffsetManager(t, 0)
	pom := initPartitionOffsetManager(t, om, coordinator, 5, "meta")

	// Error on one partition
	ocResponse := new(vendor.OffsetCommitResponse)
	ocResponse.AddError("my_topic", 0, vendor.ErrOffsetOutOfRange)
	ocResponse.AddError("my_topic", 1, vendor.ErrNoError)
	coordinator.Returns(ocResponse)

	newCoordinator := vendor.NewMockBroker(t, 3)

	// For RefreshCoordinator()
	broker.Returns(&vendor.ConsumerMetadataResponse{
		CoordinatorID:   newCoordinator.BrokerID(),
		CoordinatorHost: "127.0.0.1",
		CoordinatorPort: newCoordinator.Port(),
	})

	// Nothing in response.Errors at all
	ocResponse2 := new(vendor.OffsetCommitResponse)
	newCoordinator.Returns(ocResponse2)

	// No error, no need to refresh coordinator

	// Error on the wrong partition for this pom
	ocResponse3 := new(vendor.OffsetCommitResponse)
	ocResponse3.AddError("my_topic", 1, vendor.ErrNoError)
	newCoordinator.Returns(ocResponse3)

	// No error, no need to refresh coordinator

	// ErrUnknownTopicOrPartition/ErrNotLeaderForPartition/ErrLeaderNotAvailable block
	ocResponse4 := new(vendor.OffsetCommitResponse)
	ocResponse4.AddError("my_topic", 0, vendor.ErrUnknownTopicOrPartition)
	newCoordinator.Returns(ocResponse4)

	// For RefreshCoordinator()
	broker.Returns(&vendor.ConsumerMetadataResponse{
		CoordinatorID:   newCoordinator.BrokerID(),
		CoordinatorHost: "127.0.0.1",
		CoordinatorPort: newCoordinator.Port(),
	})

	// Normal error response
	ocResponse5 := new(vendor.OffsetCommitResponse)
	ocResponse5.AddError("my_topic", 0, vendor.ErrNoError)
	newCoordinator.Returns(ocResponse5)

	pom.MarkOffset(100, "modified_meta")

	err := pom.Close()
	if err != nil {
		t.Error(err)
	}

	broker.Close()
	coordinator.Close()
	newCoordinator.Close()
	vendor.safeClose(t, om)
	vendor.safeClose(t, testClient)
}

// Test of recovery from abort
func TestAbortPartitionOffsetManager(t *testing.T) {
	om, testClient, broker, coordinator := initOffsetManager(t, 0)
	pom := initPartitionOffsetManager(t, om, coordinator, 5, "meta")

	// this triggers an error in the CommitOffset request,
	// which leads to the abort call
	coordinator.Close()

	// Response to refresh coordinator request
	newCoordinator := vendor.NewMockBroker(t, 3)
	broker.Returns(&vendor.ConsumerMetadataResponse{
		CoordinatorID:   newCoordinator.BrokerID(),
		CoordinatorHost: "127.0.0.1",
		CoordinatorPort: newCoordinator.Port(),
	})

	ocResponse := new(vendor.OffsetCommitResponse)
	ocResponse.AddError("my_topic", 0, vendor.ErrNoError)
	newCoordinator.Returns(ocResponse)

	pom.MarkOffset(100, "modified_meta")

	vendor.safeClose(t, pom)
	vendor.safeClose(t, om)
	broker.Close()
	vendor.safeClose(t, testClient)
}
