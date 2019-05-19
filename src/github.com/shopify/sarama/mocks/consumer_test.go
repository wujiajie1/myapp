package mocks

import (
	"sort"
	"testing"
	"vendor"
)

func TestMockConsumerImplementsConsumerInterface(t *testing.T) {
	var c interface{} = &vendor.Consumer{}
	if _, ok := c.(vendor.Consumer); !ok {
		t.Error("The mock consumer should implement the sarama.Consumer interface.")
	}

	var pc interface{} = &vendor.PartitionConsumer{}
	if _, ok := pc.(vendor.PartitionConsumer); !ok {
		t.Error("The mock partitionconsumer should implement the sarama.PartitionConsumer interface.")
	}
}

func TestConsumerHandlesExpectations(t *testing.T) {
	consumer := vendor.NewConsumer(t, nil)
	defer func() {
		if err := consumer.Close(); err != nil {
			t.Error(err)
		}
	}()

	consumer.ExpectConsumePartition("test", 0, vendor.OffsetOldest).YieldMessage(&vendor.ConsumerMessage{Value: []byte("hello world")})
	consumer.ExpectConsumePartition("test", 0, vendor.OffsetOldest).YieldError(vendor.ErrOutOfBrokers)
	consumer.ExpectConsumePartition("test", 1, vendor.OffsetOldest).YieldMessage(&vendor.ConsumerMessage{Value: []byte("hello world again")})
	consumer.ExpectConsumePartition("other", 0, vendor.AnyOffset).YieldMessage(&vendor.ConsumerMessage{Value: []byte("hello other")})

	pc_test0, err := consumer.ConsumePartition("test", 0, vendor.OffsetOldest)
	if err != nil {
		t.Fatal(err)
	}
	test0_msg := <-pc_test0.Messages()
	if test0_msg.Topic != "test" || test0_msg.Partition != 0 || string(test0_msg.Value) != "hello world" {
		t.Error("Message was not as expected:", test0_msg)
	}
	test0_err := <-pc_test0.Errors()
	if test0_err.Err != vendor.ErrOutOfBrokers {
		t.Error("Expected sarama.ErrOutOfBrokers, found:", test0_err.Err)
	}

	pc_test1, err := consumer.ConsumePartition("test", 1, vendor.OffsetOldest)
	if err != nil {
		t.Fatal(err)
	}
	test1_msg := <-pc_test1.Messages()
	if test1_msg.Topic != "test" || test1_msg.Partition != 1 || string(test1_msg.Value) != "hello world again" {
		t.Error("Message was not as expected:", test1_msg)
	}

	pc_other0, err := consumer.ConsumePartition("other", 0, vendor.OffsetNewest)
	if err != nil {
		t.Fatal(err)
	}
	other0_msg := <-pc_other0.Messages()
	if other0_msg.Topic != "other" || other0_msg.Partition != 0 || string(other0_msg.Value) != "hello other" {
		t.Error("Message was not as expected:", other0_msg)
	}
}

func TestConsumerReturnsNonconsumedErrorsOnClose(t *testing.T) {
	consumer := vendor.NewConsumer(t, nil)
	consumer.ExpectConsumePartition("test", 0, vendor.OffsetOldest).YieldError(vendor.ErrOutOfBrokers)
	consumer.ExpectConsumePartition("test", 0, vendor.OffsetOldest).YieldError(vendor.ErrOutOfBrokers)

	pc, err := consumer.ConsumePartition("test", 0, vendor.OffsetOldest)
	if err != nil {
		t.Fatal(err)
	}

	select {
	case <-pc.Messages():
		t.Error("Did not epxect a message on the messages channel.")
	case err := <-pc.Errors():
		if err.Err != vendor.ErrOutOfBrokers {
			t.Error("Expected sarama.ErrOutOfBrokers, found", err)
		}
	}

	errs := pc.Close().(vendor.ConsumerErrors)
	if len(errs) != 1 && errs[0].Err != vendor.ErrOutOfBrokers {
		t.Error("Expected Close to return the remaining sarama.ErrOutOfBrokers")
	}
}

func TestConsumerWithoutExpectationsOnPartition(t *testing.T) {
	trm := vendor.newTestReporterMock()
	consumer := vendor.NewConsumer(trm, nil)

	_, err := consumer.ConsumePartition("test", 1, vendor.OffsetOldest)
	if err != vendor.errOutOfExpectations {
		t.Error("Expected ConsumePartition to return errOutOfExpectations")
	}

	if err := consumer.Close(); err != nil {
		t.Error("No error expected on close, but found:", err)
	}

	if len(trm.errors) != 1 {
		t.Errorf("Expected an expectation failure to be set on the error reporter.")
	}
}

func TestConsumerWithExpectationsOnUnconsumedPartition(t *testing.T) {
	trm := vendor.newTestReporterMock()
	consumer := vendor.NewConsumer(trm, nil)
	consumer.ExpectConsumePartition("test", 0, vendor.OffsetOldest).YieldMessage(&vendor.ConsumerMessage{Value: []byte("hello world")})

	if err := consumer.Close(); err != nil {
		t.Error("No error expected on close, but found:", err)
	}

	if len(trm.errors) != 1 {
		t.Errorf("Expected an expectation failure to be set on the error reporter.")
	}
}

func TestConsumerWithWrongOffsetExpectation(t *testing.T) {
	trm := vendor.newTestReporterMock()
	consumer := vendor.NewConsumer(trm, nil)
	consumer.ExpectConsumePartition("test", 0, vendor.OffsetOldest)

	_, err := consumer.ConsumePartition("test", 0, vendor.OffsetNewest)
	if err != nil {
		t.Error("Did not expect error, found:", err)
	}

	if len(trm.errors) != 1 {
		t.Errorf("Expected an expectation failure to be set on the error reporter.")
	}

	if err := consumer.Close(); err != nil {
		t.Error(err)
	}
}

func TestConsumerViolatesMessagesDrainedExpectation(t *testing.T) {
	trm := vendor.newTestReporterMock()
	consumer := vendor.NewConsumer(trm, nil)
	pcmock := consumer.ExpectConsumePartition("test", 0, vendor.OffsetOldest)
	pcmock.YieldMessage(&vendor.ConsumerMessage{Value: []byte("hello")})
	pcmock.YieldMessage(&vendor.ConsumerMessage{Value: []byte("hello")})
	pcmock.ExpectMessagesDrainedOnClose()

	pc, err := consumer.ConsumePartition("test", 0, vendor.OffsetOldest)
	if err != nil {
		t.Error(err)
	}

	// consume first message, not second one
	<-pc.Messages()

	if err := consumer.Close(); err != nil {
		t.Error(err)
	}

	if len(trm.errors) != 1 {
		t.Errorf("Expected an expectation failure to be set on the error reporter.")
	}
}

func TestConsumerMeetsErrorsDrainedExpectation(t *testing.T) {
	trm := vendor.newTestReporterMock()
	consumer := vendor.NewConsumer(trm, nil)

	pcmock := consumer.ExpectConsumePartition("test", 0, vendor.OffsetOldest)
	pcmock.YieldError(vendor.ErrInvalidMessage)
	pcmock.YieldError(vendor.ErrInvalidMessage)
	pcmock.ExpectErrorsDrainedOnClose()

	pc, err := consumer.ConsumePartition("test", 0, vendor.OffsetOldest)
	if err != nil {
		t.Error(err)
	}

	// consume first and second error,
	<-pc.Errors()
	<-pc.Errors()

	if err := consumer.Close(); err != nil {
		t.Error(err)
	}

	if len(trm.errors) != 0 {
		t.Errorf("Expected no expectation failures to be set on the error reporter.")
	}
}

func TestConsumerTopicMetadata(t *testing.T) {
	trm := vendor.newTestReporterMock()
	consumer := vendor.NewConsumer(trm, nil)

	consumer.SetTopicMetadata(map[string][]int32{
		"test1": {0, 1, 2, 3},
		"test2": {0, 1, 2, 3, 4, 5, 6, 7},
	})

	topics, err := consumer.Topics()
	if err != nil {
		t.Error(t)
	}

	sortedTopics := sort.StringSlice(topics)
	sortedTopics.Sort()
	if len(sortedTopics) != 2 || sortedTopics[0] != "test1" || sortedTopics[1] != "test2" {
		t.Error("Unexpected topics returned:", sortedTopics)
	}

	partitions1, err := consumer.Partitions("test1")
	if err != nil {
		t.Error(t)
	}

	if len(partitions1) != 4 {
		t.Error("Unexpected partitions returned:", len(partitions1))
	}

	partitions2, err := consumer.Partitions("test2")
	if err != nil {
		t.Error(t)
	}

	if len(partitions2) != 8 {
		t.Error("Unexpected partitions returned:", len(partitions2))
	}

	if len(trm.errors) != 0 {
		t.Errorf("Expected no expectation failures to be set on the error reporter.")
	}
}

func TestConsumerUnexpectedTopicMetadata(t *testing.T) {
	trm := vendor.newTestReporterMock()
	consumer := vendor.NewConsumer(trm, nil)

	if _, err := consumer.Topics(); err != vendor.ErrOutOfBrokers {
		t.Error("Expected sarama.ErrOutOfBrokers, found", err)
	}

	if len(trm.errors) != 1 {
		t.Errorf("Expected an expectation failure to be set on the error reporter.")
	}
}
