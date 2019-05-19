package sarama

import (
	"errors"
	"log"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"vendor"
)

const TestMessage = "ABC THE MESSAGE"

func closeProducer(t *testing.T, p vendor.AsyncProducer) {
	var wg sync.WaitGroup
	p.AsyncClose()

	wg.Add(2)
	go func() {
		for range p.Successes() {
			t.Error("Unexpected message on Successes()")
		}
		wg.Done()
	}()
	go func() {
		for msg := range p.Errors() {
			t.Error(msg.Err)
		}
		wg.Done()
	}()
	wg.Wait()
}

func expectResults(t *testing.T, p vendor.AsyncProducer, successes, errors int) {
	expect := successes + errors
	for expect > 0 {
		select {
		case msg := <-p.Errors():
			if msg.Msg.flags != 0 {
				t.Error("Message had flags set")
			}
			errors--
			expect--
			if errors < 0 {
				t.Error(msg.Err)
			}
		case msg := <-p.Successes():
			if msg.flags != 0 {
				t.Error("Message had flags set")
			}
			successes--
			expect--
			if successes < 0 {
				t.Error("Too many successes")
			}
		}
	}
	if successes != 0 || errors != 0 {
		t.Error("Unexpected successes", successes, "or errors", errors)
	}
}

type testPartitioner chan *int32

func (p testPartitioner) Partition(msg *vendor.ProducerMessage, numPartitions int32) (int32, error) {
	part := <-p
	if part == nil {
		return 0, errors.New("BOOM")
	}

	return *part, nil
}

func (p testPartitioner) RequiresConsistency() bool {
	return true
}

func (p testPartitioner) feed(partition int32) {
	p <- &partition
}

type flakyEncoder bool

func (f flakyEncoder) Length() int {
	return len(TestMessage)
}

func (f flakyEncoder) Encode() ([]byte, error) {
	if !bool(f) {
		return nil, errors.New("flaky encoding error")
	}
	return []byte(TestMessage), nil
}

func TestAsyncProducer(t *testing.T) {
	seedBroker := vendor.NewMockBroker(t, 1)
	leader := vendor.NewMockBroker(t, 2)

	metadataResponse := new(vendor.MetadataResponse)
	metadataResponse.AddBroker(leader.Addr(), leader.BrokerID())
	metadataResponse.AddTopicPartition("my_topic", 0, leader.BrokerID(), nil, nil, nil, vendor.ErrNoError)
	seedBroker.Returns(metadataResponse)

	prodSuccess := new(vendor.ProduceResponse)
	prodSuccess.AddTopicPartition("my_topic", 0, vendor.ErrNoError)
	leader.Returns(prodSuccess)

	config := vendor.NewConfig()
	config.Producer.Flush.Messages = 10
	config.Producer.Return.Successes = true
	producer, err := vendor.NewAsyncProducer([]string{seedBroker.Addr()}, config)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 10; i++ {
		producer.Input() <- &vendor.ProducerMessage{Topic: "my_topic", Key: nil, Value: vendor.StringEncoder(TestMessage), Metadata: i}
	}
	for i := 0; i < 10; i++ {
		select {
		case msg := <-producer.Errors():
			t.Error(msg.Err)
			if msg.Msg.flags != 0 {
				t.Error("Message had flags set")
			}
		case msg := <-producer.Successes():
			if msg.flags != 0 {
				t.Error("Message had flags set")
			}
			if msg.Metadata.(int) != i {
				t.Error("Message metadata did not match")
			}
		case <-time.After(time.Second):
			t.Errorf("Timeout waiting for msg #%d", i)
			goto done
		}
	}
done:
	closeProducer(t, producer)
	leader.Close()
	seedBroker.Close()
}

func TestAsyncProducerMultipleFlushes(t *testing.T) {
	seedBroker := vendor.NewMockBroker(t, 1)
	leader := vendor.NewMockBroker(t, 2)

	metadataResponse := new(vendor.MetadataResponse)
	metadataResponse.AddBroker(leader.Addr(), leader.BrokerID())
	metadataResponse.AddTopicPartition("my_topic", 0, leader.BrokerID(), nil, nil, nil, vendor.ErrNoError)
	seedBroker.Returns(metadataResponse)

	prodSuccess := new(vendor.ProduceResponse)
	prodSuccess.AddTopicPartition("my_topic", 0, vendor.ErrNoError)
	leader.Returns(prodSuccess)
	leader.Returns(prodSuccess)
	leader.Returns(prodSuccess)

	config := vendor.NewConfig()
	config.Producer.Flush.Messages = 5
	config.Producer.Return.Successes = true
	producer, err := vendor.NewAsyncProducer([]string{seedBroker.Addr()}, config)
	if err != nil {
		t.Fatal(err)
	}

	for flush := 0; flush < 3; flush++ {
		for i := 0; i < 5; i++ {
			producer.Input() <- &vendor.ProducerMessage{Topic: "my_topic", Key: nil, Value: vendor.StringEncoder(TestMessage)}
		}
		expectResults(t, producer, 5, 0)
	}

	closeProducer(t, producer)
	leader.Close()
	seedBroker.Close()
}

func TestAsyncProducerMultipleBrokers(t *testing.T) {
	seedBroker := vendor.NewMockBroker(t, 1)
	leader0 := vendor.NewMockBroker(t, 2)
	leader1 := vendor.NewMockBroker(t, 3)

	metadataResponse := new(vendor.MetadataResponse)
	metadataResponse.AddBroker(leader0.Addr(), leader0.BrokerID())
	metadataResponse.AddBroker(leader1.Addr(), leader1.BrokerID())
	metadataResponse.AddTopicPartition("my_topic", 0, leader0.BrokerID(), nil, nil, nil, vendor.ErrNoError)
	metadataResponse.AddTopicPartition("my_topic", 1, leader1.BrokerID(), nil, nil, nil, vendor.ErrNoError)
	seedBroker.Returns(metadataResponse)

	prodResponse0 := new(vendor.ProduceResponse)
	prodResponse0.AddTopicPartition("my_topic", 0, vendor.ErrNoError)
	leader0.Returns(prodResponse0)

	prodResponse1 := new(vendor.ProduceResponse)
	prodResponse1.AddTopicPartition("my_topic", 1, vendor.ErrNoError)
	leader1.Returns(prodResponse1)

	config := vendor.NewConfig()
	config.Producer.Flush.Messages = 5
	config.Producer.Return.Successes = true
	config.Producer.Partitioner = vendor.NewRoundRobinPartitioner
	producer, err := vendor.NewAsyncProducer([]string{seedBroker.Addr()}, config)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 10; i++ {
		producer.Input() <- &vendor.ProducerMessage{Topic: "my_topic", Key: nil, Value: vendor.StringEncoder(TestMessage)}
	}
	expectResults(t, producer, 10, 0)

	closeProducer(t, producer)
	leader1.Close()
	leader0.Close()
	seedBroker.Close()
}

func TestAsyncProducerCustomPartitioner(t *testing.T) {
	seedBroker := vendor.NewMockBroker(t, 1)
	leader := vendor.NewMockBroker(t, 2)

	metadataResponse := new(vendor.MetadataResponse)
	metadataResponse.AddBroker(leader.Addr(), leader.BrokerID())
	metadataResponse.AddTopicPartition("my_topic", 0, leader.BrokerID(), nil, nil, nil, vendor.ErrNoError)
	seedBroker.Returns(metadataResponse)

	prodResponse := new(vendor.ProduceResponse)
	prodResponse.AddTopicPartition("my_topic", 0, vendor.ErrNoError)
	leader.Returns(prodResponse)

	config := vendor.NewConfig()
	config.Producer.Flush.Messages = 2
	config.Producer.Return.Successes = true
	config.Producer.Partitioner = func(topic string) vendor.Partitioner {
		p := make(testPartitioner)
		go func() {
			p.feed(0)
			p <- nil
			p <- nil
			p <- nil
			p.feed(0)
		}()
		return p
	}
	producer, err := vendor.NewAsyncProducer([]string{seedBroker.Addr()}, config)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 5; i++ {
		producer.Input() <- &vendor.ProducerMessage{Topic: "my_topic", Key: nil, Value: vendor.StringEncoder(TestMessage)}
	}
	expectResults(t, producer, 2, 3)

	closeProducer(t, producer)
	leader.Close()
	seedBroker.Close()
}

func TestAsyncProducerFailureRetry(t *testing.T) {
	seedBroker := vendor.NewMockBroker(t, 1)
	leader1 := vendor.NewMockBroker(t, 2)
	leader2 := vendor.NewMockBroker(t, 3)

	metadataLeader1 := new(vendor.MetadataResponse)
	metadataLeader1.AddBroker(leader1.Addr(), leader1.BrokerID())
	metadataLeader1.AddTopicPartition("my_topic", 0, leader1.BrokerID(), nil, nil, nil, vendor.ErrNoError)
	seedBroker.Returns(metadataLeader1)

	config := vendor.NewConfig()
	config.Producer.Flush.Messages = 10
	config.Producer.Return.Successes = true
	config.Producer.Retry.Backoff = 0
	producer, err := vendor.NewAsyncProducer([]string{seedBroker.Addr()}, config)
	if err != nil {
		t.Fatal(err)
	}
	seedBroker.Close()

	for i := 0; i < 10; i++ {
		producer.Input() <- &vendor.ProducerMessage{Topic: "my_topic", Key: nil, Value: vendor.StringEncoder(TestMessage)}
	}
	prodNotLeader := new(vendor.ProduceResponse)
	prodNotLeader.AddTopicPartition("my_topic", 0, vendor.ErrNotLeaderForPartition)
	leader1.Returns(prodNotLeader)

	metadataLeader2 := new(vendor.MetadataResponse)
	metadataLeader2.AddBroker(leader2.Addr(), leader2.BrokerID())
	metadataLeader2.AddTopicPartition("my_topic", 0, leader2.BrokerID(), nil, nil, nil, vendor.ErrNoError)
	leader1.Returns(metadataLeader2)

	prodSuccess := new(vendor.ProduceResponse)
	prodSuccess.AddTopicPartition("my_topic", 0, vendor.ErrNoError)
	leader2.Returns(prodSuccess)
	expectResults(t, producer, 10, 0)
	leader1.Close()

	for i := 0; i < 10; i++ {
		producer.Input() <- &vendor.ProducerMessage{Topic: "my_topic", Key: nil, Value: vendor.StringEncoder(TestMessage)}
	}
	leader2.Returns(prodSuccess)
	expectResults(t, producer, 10, 0)

	leader2.Close()
	closeProducer(t, producer)
}

func TestAsyncProducerRecoveryWithRetriesDisabled(t *testing.T) {

	tt := func(t *testing.T, kErr vendor.KError) {
		seedBroker := vendor.NewMockBroker(t, 1)
		leader1 := vendor.NewMockBroker(t, 2)
		leader2 := vendor.NewMockBroker(t, 3)

		metadataLeader1 := new(vendor.MetadataResponse)
		metadataLeader1.AddBroker(leader1.Addr(), leader1.BrokerID())
		metadataLeader1.AddTopicPartition("my_topic", 0, leader1.BrokerID(), nil, nil, nil, vendor.ErrNoError)
		metadataLeader1.AddTopicPartition("my_topic", 1, leader1.BrokerID(), nil, nil, nil, vendor.ErrNoError)
		seedBroker.Returns(metadataLeader1)

		config := vendor.NewConfig()
		config.Producer.Flush.Messages = 2
		config.Producer.Return.Successes = true
		config.Producer.Retry.Max = 0 // disable!
		config.Producer.Retry.Backoff = 0
		config.Producer.Partitioner = vendor.NewManualPartitioner
		producer, err := vendor.NewAsyncProducer([]string{seedBroker.Addr()}, config)
		if err != nil {
			t.Fatal(err)
		}
		seedBroker.Close()

		producer.Input() <- &vendor.ProducerMessage{Topic: "my_topic", Key: nil, Value: vendor.StringEncoder(TestMessage), Partition: 0}
		producer.Input() <- &vendor.ProducerMessage{Topic: "my_topic", Key: nil, Value: vendor.StringEncoder(TestMessage), Partition: 1}
		prodNotLeader := new(vendor.ProduceResponse)
		prodNotLeader.AddTopicPartition("my_topic", 0, kErr)
		prodNotLeader.AddTopicPartition("my_topic", 1, kErr)
		leader1.Returns(prodNotLeader)
		expectResults(t, producer, 0, 2)

		producer.Input() <- &vendor.ProducerMessage{Topic: "my_topic", Key: nil, Value: vendor.StringEncoder(TestMessage), Partition: 0}
		metadataLeader2 := new(vendor.MetadataResponse)
		metadataLeader2.AddBroker(leader2.Addr(), leader2.BrokerID())
		metadataLeader2.AddTopicPartition("my_topic", 0, leader2.BrokerID(), nil, nil, nil, vendor.ErrNoError)
		metadataLeader2.AddTopicPartition("my_topic", 1, leader2.BrokerID(), nil, nil, nil, vendor.ErrNoError)
		leader1.Returns(metadataLeader2)
		leader1.Returns(metadataLeader2)

		producer.Input() <- &vendor.ProducerMessage{Topic: "my_topic", Key: nil, Value: vendor.StringEncoder(TestMessage), Partition: 1}
		prodSuccess := new(vendor.ProduceResponse)
		prodSuccess.AddTopicPartition("my_topic", 0, vendor.ErrNoError)
		prodSuccess.AddTopicPartition("my_topic", 1, vendor.ErrNoError)
		leader2.Returns(prodSuccess)
		expectResults(t, producer, 2, 0)

		leader1.Close()
		leader2.Close()
		closeProducer(t, producer)
	}

	t.Run("retriable error", func(t *testing.T) {
		tt(t, vendor.ErrNotLeaderForPartition)
	})

	t.Run("non-retriable error", func(t *testing.T) {
		tt(t, vendor.ErrNotController)
	})
}

func TestAsyncProducerEncoderFailures(t *testing.T) {
	seedBroker := vendor.NewMockBroker(t, 1)
	leader := vendor.NewMockBroker(t, 2)

	metadataResponse := new(vendor.MetadataResponse)
	metadataResponse.AddBroker(leader.Addr(), leader.BrokerID())
	metadataResponse.AddTopicPartition("my_topic", 0, leader.BrokerID(), nil, nil, nil, vendor.ErrNoError)
	seedBroker.Returns(metadataResponse)

	prodSuccess := new(vendor.ProduceResponse)
	prodSuccess.AddTopicPartition("my_topic", 0, vendor.ErrNoError)
	leader.Returns(prodSuccess)
	leader.Returns(prodSuccess)
	leader.Returns(prodSuccess)

	config := vendor.NewConfig()
	config.Producer.Flush.Messages = 1
	config.Producer.Return.Successes = true
	config.Producer.Partitioner = vendor.NewManualPartitioner
	producer, err := vendor.NewAsyncProducer([]string{seedBroker.Addr()}, config)
	if err != nil {
		t.Fatal(err)
	}

	for flush := 0; flush < 3; flush++ {
		producer.Input() <- &vendor.ProducerMessage{Topic: "my_topic", Key: flakyEncoder(true), Value: flakyEncoder(false)}
		producer.Input() <- &vendor.ProducerMessage{Topic: "my_topic", Key: flakyEncoder(false), Value: flakyEncoder(true)}
		producer.Input() <- &vendor.ProducerMessage{Topic: "my_topic", Key: flakyEncoder(true), Value: flakyEncoder(true)}
		expectResults(t, producer, 1, 2)
	}

	closeProducer(t, producer)
	leader.Close()
	seedBroker.Close()
}

// If a Kafka broker becomes unavailable and then returns back in service, then
// producer reconnects to it and continues sending messages.
func TestAsyncProducerBrokerBounce(t *testing.T) {
	// Given
	seedBroker := vendor.NewMockBroker(t, 1)
	leader := vendor.NewMockBroker(t, 2)
	leaderAddr := leader.Addr()

	metadataResponse := new(vendor.MetadataResponse)
	metadataResponse.AddBroker(leaderAddr, leader.BrokerID())
	metadataResponse.AddTopicPartition("my_topic", 0, leader.BrokerID(), nil, nil, nil, vendor.ErrNoError)
	seedBroker.Returns(metadataResponse)

	prodSuccess := new(vendor.ProduceResponse)
	prodSuccess.AddTopicPartition("my_topic", 0, vendor.ErrNoError)

	config := vendor.NewConfig()
	config.Producer.Flush.Messages = 1
	config.Producer.Return.Successes = true
	config.Producer.Retry.Backoff = 0
	producer, err := vendor.NewAsyncProducer([]string{seedBroker.Addr()}, config)
	if err != nil {
		t.Fatal(err)
	}
	producer.Input() <- &vendor.ProducerMessage{Topic: "my_topic", Key: nil, Value: vendor.StringEncoder(TestMessage)}
	leader.Returns(prodSuccess)
	expectResults(t, producer, 1, 0)

	// When: a broker connection gets reset by a broker (network glitch, restart, you name it).
	leader.Close()                                      // producer should get EOF
	leader = vendor.NewMockBrokerAddr(t, 2, leaderAddr) // start it up again right away for giggles
	seedBroker.Returns(metadataResponse)                // tell it to go to broker 2 again

	// Then: a produced message goes through the new broker connection.
	producer.Input() <- &vendor.ProducerMessage{Topic: "my_topic", Key: nil, Value: vendor.StringEncoder(TestMessage)}
	leader.Returns(prodSuccess)
	expectResults(t, producer, 1, 0)

	closeProducer(t, producer)
	seedBroker.Close()
	leader.Close()
}

func TestAsyncProducerBrokerBounceWithStaleMetadata(t *testing.T) {
	seedBroker := vendor.NewMockBroker(t, 1)
	leader1 := vendor.NewMockBroker(t, 2)
	leader2 := vendor.NewMockBroker(t, 3)

	metadataLeader1 := new(vendor.MetadataResponse)
	metadataLeader1.AddBroker(leader1.Addr(), leader1.BrokerID())
	metadataLeader1.AddTopicPartition("my_topic", 0, leader1.BrokerID(), nil, nil, nil, vendor.ErrNoError)
	seedBroker.Returns(metadataLeader1)

	config := vendor.NewConfig()
	config.Producer.Flush.Messages = 10
	config.Producer.Return.Successes = true
	config.Producer.Retry.Max = 3
	config.Producer.Retry.Backoff = 0
	producer, err := vendor.NewAsyncProducer([]string{seedBroker.Addr()}, config)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 10; i++ {
		producer.Input() <- &vendor.ProducerMessage{Topic: "my_topic", Key: nil, Value: vendor.StringEncoder(TestMessage)}
	}
	leader1.Close()                     // producer should get EOF
	seedBroker.Returns(metadataLeader1) // tell it to go to leader1 again even though it's still down
	seedBroker.Returns(metadataLeader1) // tell it to go to leader1 again even though it's still down

	// ok fine, tell it to go to leader2 finally
	metadataLeader2 := new(vendor.MetadataResponse)
	metadataLeader2.AddBroker(leader2.Addr(), leader2.BrokerID())
	metadataLeader2.AddTopicPartition("my_topic", 0, leader2.BrokerID(), nil, nil, nil, vendor.ErrNoError)
	seedBroker.Returns(metadataLeader2)

	prodSuccess := new(vendor.ProduceResponse)
	prodSuccess.AddTopicPartition("my_topic", 0, vendor.ErrNoError)
	leader2.Returns(prodSuccess)
	expectResults(t, producer, 10, 0)
	seedBroker.Close()
	leader2.Close()

	closeProducer(t, producer)
}

func TestAsyncProducerMultipleRetries(t *testing.T) {
	seedBroker := vendor.NewMockBroker(t, 1)
	leader1 := vendor.NewMockBroker(t, 2)
	leader2 := vendor.NewMockBroker(t, 3)

	metadataLeader1 := new(vendor.MetadataResponse)
	metadataLeader1.AddBroker(leader1.Addr(), leader1.BrokerID())
	metadataLeader1.AddTopicPartition("my_topic", 0, leader1.BrokerID(), nil, nil, nil, vendor.ErrNoError)
	seedBroker.Returns(metadataLeader1)

	config := vendor.NewConfig()
	config.Producer.Flush.Messages = 10
	config.Producer.Return.Successes = true
	config.Producer.Retry.Max = 4
	config.Producer.Retry.Backoff = 0
	producer, err := vendor.NewAsyncProducer([]string{seedBroker.Addr()}, config)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 10; i++ {
		producer.Input() <- &vendor.ProducerMessage{Topic: "my_topic", Key: nil, Value: vendor.StringEncoder(TestMessage)}
	}
	prodNotLeader := new(vendor.ProduceResponse)
	prodNotLeader.AddTopicPartition("my_topic", 0, vendor.ErrNotLeaderForPartition)
	leader1.Returns(prodNotLeader)

	metadataLeader2 := new(vendor.MetadataResponse)
	metadataLeader2.AddBroker(leader2.Addr(), leader2.BrokerID())
	metadataLeader2.AddTopicPartition("my_topic", 0, leader2.BrokerID(), nil, nil, nil, vendor.ErrNoError)

	seedBroker.Returns(metadataLeader2)
	leader2.Returns(prodNotLeader)
	seedBroker.Returns(metadataLeader1)
	leader1.Returns(prodNotLeader)
	seedBroker.Returns(metadataLeader1)
	leader1.Returns(prodNotLeader)
	seedBroker.Returns(metadataLeader2)

	prodSuccess := new(vendor.ProduceResponse)
	prodSuccess.AddTopicPartition("my_topic", 0, vendor.ErrNoError)
	leader2.Returns(prodSuccess)
	expectResults(t, producer, 10, 0)

	for i := 0; i < 10; i++ {
		producer.Input() <- &vendor.ProducerMessage{Topic: "my_topic", Key: nil, Value: vendor.StringEncoder(TestMessage)}
	}
	leader2.Returns(prodSuccess)
	expectResults(t, producer, 10, 0)

	seedBroker.Close()
	leader1.Close()
	leader2.Close()
	closeProducer(t, producer)
}

func TestAsyncProducerMultipleRetriesWithBackoffFunc(t *testing.T) {
	seedBroker := vendor.NewMockBroker(t, 1)
	leader1 := vendor.NewMockBroker(t, 2)
	leader2 := vendor.NewMockBroker(t, 3)

	metadataLeader1 := new(vendor.MetadataResponse)
	metadataLeader1.AddBroker(leader1.Addr(), leader1.BrokerID())
	metadataLeader1.AddTopicPartition("my_topic", 0, leader1.BrokerID(), nil, nil, nil, vendor.ErrNoError)
	seedBroker.Returns(metadataLeader1)

	config := vendor.NewConfig()
	config.Producer.Flush.Messages = 1
	config.Producer.Return.Successes = true
	config.Producer.Retry.Max = 4

	backoffCalled := make([]int32, config.Producer.Retry.Max+1)
	config.Producer.Retry.BackoffFunc = func(retries, maxRetries int) time.Duration {
		atomic.AddInt32(&backoffCalled[retries-1], 1)
		return 0
	}
	producer, err := vendor.NewAsyncProducer([]string{seedBroker.Addr()}, config)
	if err != nil {
		t.Fatal(err)
	}

	producer.Input() <- &vendor.ProducerMessage{Topic: "my_topic", Key: nil, Value: vendor.StringEncoder(TestMessage)}
	prodNotLeader := new(vendor.ProduceResponse)
	prodNotLeader.AddTopicPartition("my_topic", 0, vendor.ErrNotLeaderForPartition)

	prodSuccess := new(vendor.ProduceResponse)
	prodSuccess.AddTopicPartition("my_topic", 0, vendor.ErrNoError)

	metadataLeader2 := new(vendor.MetadataResponse)
	metadataLeader2.AddBroker(leader2.Addr(), leader2.BrokerID())
	metadataLeader2.AddTopicPartition("my_topic", 0, leader2.BrokerID(), nil, nil, nil, vendor.ErrNoError)

	leader1.Returns(prodNotLeader)
	seedBroker.Returns(metadataLeader2)
	leader2.Returns(prodNotLeader)
	seedBroker.Returns(metadataLeader1)
	leader1.Returns(prodNotLeader)
	seedBroker.Returns(metadataLeader1)
	leader1.Returns(prodNotLeader)
	seedBroker.Returns(metadataLeader2)
	leader2.Returns(prodSuccess)

	expectResults(t, producer, 1, 0)

	producer.Input() <- &vendor.ProducerMessage{Topic: "my_topic", Key: nil, Value: vendor.StringEncoder(TestMessage)}
	leader2.Returns(prodSuccess)
	expectResults(t, producer, 1, 0)

	seedBroker.Close()
	leader1.Close()
	leader2.Close()
	closeProducer(t, producer)

	for i := 0; i < config.Producer.Retry.Max; i++ {
		if atomic.LoadInt32(&backoffCalled[i]) != 1 {
			t.Errorf("expected one retry attempt #%d", i)
		}
	}
	if atomic.LoadInt32(&backoffCalled[config.Producer.Retry.Max]) != 0 {
		t.Errorf("expected no retry attempt #%d", config.Producer.Retry.Max)
	}
}

func TestAsyncProducerOutOfRetries(t *testing.T) {
	t.Skip("Enable once bug #294 is fixed.")

	seedBroker := vendor.NewMockBroker(t, 1)
	leader := vendor.NewMockBroker(t, 2)

	metadataResponse := new(vendor.MetadataResponse)
	metadataResponse.AddBroker(leader.Addr(), leader.BrokerID())
	metadataResponse.AddTopicPartition("my_topic", 0, leader.BrokerID(), nil, nil, nil, vendor.ErrNoError)
	seedBroker.Returns(metadataResponse)

	config := vendor.NewConfig()
	config.Producer.Flush.Messages = 10
	config.Producer.Return.Successes = true
	config.Producer.Retry.Backoff = 0
	config.Producer.Retry.Max = 0
	producer, err := vendor.NewAsyncProducer([]string{seedBroker.Addr()}, config)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 10; i++ {
		producer.Input() <- &vendor.ProducerMessage{Topic: "my_topic", Key: nil, Value: vendor.StringEncoder(TestMessage)}
	}

	prodNotLeader := new(vendor.ProduceResponse)
	prodNotLeader.AddTopicPartition("my_topic", 0, vendor.ErrNotLeaderForPartition)
	leader.Returns(prodNotLeader)

	for i := 0; i < 10; i++ {
		select {
		case msg := <-producer.Errors():
			if msg.Err != vendor.ErrNotLeaderForPartition {
				t.Error(msg.Err)
			}
		case <-producer.Successes():
			t.Error("Unexpected success")
		}
	}

	seedBroker.Returns(metadataResponse)

	for i := 0; i < 10; i++ {
		producer.Input() <- &vendor.ProducerMessage{Topic: "my_topic", Key: nil, Value: vendor.StringEncoder(TestMessage)}
	}

	prodSuccess := new(vendor.ProduceResponse)
	prodSuccess.AddTopicPartition("my_topic", 0, vendor.ErrNoError)
	leader.Returns(prodSuccess)

	expectResults(t, producer, 10, 0)

	leader.Close()
	seedBroker.Close()
	vendor.safeClose(t, producer)
}

func TestAsyncProducerRetryWithReferenceOpen(t *testing.T) {
	seedBroker := vendor.NewMockBroker(t, 1)
	leader := vendor.NewMockBroker(t, 2)
	leaderAddr := leader.Addr()

	metadataResponse := new(vendor.MetadataResponse)
	metadataResponse.AddBroker(leaderAddr, leader.BrokerID())
	metadataResponse.AddTopicPartition("my_topic", 0, leader.BrokerID(), nil, nil, nil, vendor.ErrNoError)
	metadataResponse.AddTopicPartition("my_topic", 1, leader.BrokerID(), nil, nil, nil, vendor.ErrNoError)
	seedBroker.Returns(metadataResponse)

	config := vendor.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Retry.Backoff = 0
	config.Producer.Retry.Max = 1
	config.Producer.Partitioner = vendor.NewRoundRobinPartitioner
	producer, err := vendor.NewAsyncProducer([]string{seedBroker.Addr()}, config)
	if err != nil {
		t.Fatal(err)
	}

	// prime partition 0
	producer.Input() <- &vendor.ProducerMessage{Topic: "my_topic", Key: nil, Value: vendor.StringEncoder(TestMessage)}
	prodSuccess := new(vendor.ProduceResponse)
	prodSuccess.AddTopicPartition("my_topic", 0, vendor.ErrNoError)
	leader.Returns(prodSuccess)
	expectResults(t, producer, 1, 0)

	// prime partition 1
	producer.Input() <- &vendor.ProducerMessage{Topic: "my_topic", Key: nil, Value: vendor.StringEncoder(TestMessage)}
	prodSuccess = new(vendor.ProduceResponse)
	prodSuccess.AddTopicPartition("my_topic", 1, vendor.ErrNoError)
	leader.Returns(prodSuccess)
	expectResults(t, producer, 1, 0)

	// reboot the broker (the producer will get EOF on its existing connection)
	leader.Close()
	leader = vendor.NewMockBrokerAddr(t, 2, leaderAddr)

	// send another message on partition 0 to trigger the EOF and retry
	producer.Input() <- &vendor.ProducerMessage{Topic: "my_topic", Key: nil, Value: vendor.StringEncoder(TestMessage)}

	// tell partition 0 to go to that broker again
	seedBroker.Returns(metadataResponse)

	// succeed this time
	prodSuccess = new(vendor.ProduceResponse)
	prodSuccess.AddTopicPartition("my_topic", 0, vendor.ErrNoError)
	leader.Returns(prodSuccess)
	expectResults(t, producer, 1, 0)

	// shutdown
	closeProducer(t, producer)
	seedBroker.Close()
	leader.Close()
}

func TestAsyncProducerFlusherRetryCondition(t *testing.T) {
	seedBroker := vendor.NewMockBroker(t, 1)
	leader := vendor.NewMockBroker(t, 2)

	metadataResponse := new(vendor.MetadataResponse)
	metadataResponse.AddBroker(leader.Addr(), leader.BrokerID())
	metadataResponse.AddTopicPartition("my_topic", 0, leader.BrokerID(), nil, nil, nil, vendor.ErrNoError)
	metadataResponse.AddTopicPartition("my_topic", 1, leader.BrokerID(), nil, nil, nil, vendor.ErrNoError)
	seedBroker.Returns(metadataResponse)

	config := vendor.NewConfig()
	config.Producer.Flush.Messages = 5
	config.Producer.Return.Successes = true
	config.Producer.Retry.Backoff = 0
	config.Producer.Retry.Max = 1
	config.Producer.Partitioner = vendor.NewManualPartitioner
	producer, err := vendor.NewAsyncProducer([]string{seedBroker.Addr()}, config)
	if err != nil {
		t.Fatal(err)
	}

	// prime partitions
	for p := int32(0); p < 2; p++ {
		for i := 0; i < 5; i++ {
			producer.Input() <- &vendor.ProducerMessage{Topic: "my_topic", Key: nil, Value: vendor.StringEncoder(TestMessage), Partition: p}
		}
		prodSuccess := new(vendor.ProduceResponse)
		prodSuccess.AddTopicPartition("my_topic", p, vendor.ErrNoError)
		leader.Returns(prodSuccess)
		expectResults(t, producer, 5, 0)
	}

	// send more messages on partition 0
	for i := 0; i < 5; i++ {
		producer.Input() <- &vendor.ProducerMessage{Topic: "my_topic", Key: nil, Value: vendor.StringEncoder(TestMessage), Partition: 0}
	}
	prodNotLeader := new(vendor.ProduceResponse)
	prodNotLeader.AddTopicPartition("my_topic", 0, vendor.ErrNotLeaderForPartition)
	leader.Returns(prodNotLeader)

	time.Sleep(50 * time.Millisecond)

	leader.SetHandlerByMap(map[string]vendor.MockResponse{
		"ProduceRequest": vendor.NewMockProduceResponse(t).
			SetVersion(0).
			SetError("my_topic", 0, vendor.ErrNoError),
	})

	// tell partition 0 to go to that broker again
	seedBroker.Returns(metadataResponse)

	// succeed this time
	expectResults(t, producer, 5, 0)

	// put five more through
	for i := 0; i < 5; i++ {
		producer.Input() <- &vendor.ProducerMessage{Topic: "my_topic", Key: nil, Value: vendor.StringEncoder(TestMessage), Partition: 0}
	}
	expectResults(t, producer, 5, 0)

	// shutdown
	closeProducer(t, producer)
	seedBroker.Close()
	leader.Close()
}

func TestAsyncProducerRetryShutdown(t *testing.T) {
	seedBroker := vendor.NewMockBroker(t, 1)
	leader := vendor.NewMockBroker(t, 2)

	metadataLeader := new(vendor.MetadataResponse)
	metadataLeader.AddBroker(leader.Addr(), leader.BrokerID())
	metadataLeader.AddTopicPartition("my_topic", 0, leader.BrokerID(), nil, nil, nil, vendor.ErrNoError)
	seedBroker.Returns(metadataLeader)

	config := vendor.NewConfig()
	config.Producer.Flush.Messages = 10
	config.Producer.Return.Successes = true
	config.Producer.Retry.Backoff = 0
	producer, err := vendor.NewAsyncProducer([]string{seedBroker.Addr()}, config)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 10; i++ {
		producer.Input() <- &vendor.ProducerMessage{Topic: "my_topic", Key: nil, Value: vendor.StringEncoder(TestMessage)}
	}
	producer.AsyncClose()
	time.Sleep(5 * time.Millisecond) // let the shutdown goroutine kick in

	producer.Input() <- &vendor.ProducerMessage{Topic: "FOO"}
	if err := <-producer.Errors(); err.Err != vendor.ErrShuttingDown {
		t.Error(err)
	}

	prodNotLeader := new(vendor.ProduceResponse)
	prodNotLeader.AddTopicPartition("my_topic", 0, vendor.ErrNotLeaderForPartition)
	leader.Returns(prodNotLeader)

	seedBroker.Returns(metadataLeader)

	prodSuccess := new(vendor.ProduceResponse)
	prodSuccess.AddTopicPartition("my_topic", 0, vendor.ErrNoError)
	leader.Returns(prodSuccess)
	expectResults(t, producer, 10, 0)

	seedBroker.Close()
	leader.Close()

	// wait for the async-closed producer to shut down fully
	for err := range producer.Errors() {
		t.Error(err)
	}
}

func TestAsyncProducerNoReturns(t *testing.T) {
	seedBroker := vendor.NewMockBroker(t, 1)
	leader := vendor.NewMockBroker(t, 2)

	metadataLeader := new(vendor.MetadataResponse)
	metadataLeader.AddBroker(leader.Addr(), leader.BrokerID())
	metadataLeader.AddTopicPartition("my_topic", 0, leader.BrokerID(), nil, nil, nil, vendor.ErrNoError)
	seedBroker.Returns(metadataLeader)

	config := vendor.NewConfig()
	config.Producer.Flush.Messages = 10
	config.Producer.Return.Successes = false
	config.Producer.Return.Errors = false
	config.Producer.Retry.Backoff = 0
	producer, err := vendor.NewAsyncProducer([]string{seedBroker.Addr()}, config)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 10; i++ {
		producer.Input() <- &vendor.ProducerMessage{Topic: "my_topic", Key: nil, Value: vendor.StringEncoder(TestMessage)}
	}

	wait := make(chan bool)
	go func() {
		if err := producer.Close(); err != nil {
			t.Error(err)
		}
		close(wait)
	}()

	prodSuccess := new(vendor.ProduceResponse)
	prodSuccess.AddTopicPartition("my_topic", 0, vendor.ErrNoError)
	leader.Returns(prodSuccess)

	<-wait
	seedBroker.Close()
	leader.Close()
}

func TestAsyncProducerIdempotentGoldenPath(t *testing.T) {
	broker := vendor.NewMockBroker(t, 1)

	metadataResponse := &vendor.MetadataResponse{
		Version:      1,
		ControllerID: 1,
	}
	metadataResponse.AddBroker(broker.Addr(), broker.BrokerID())
	metadataResponse.AddTopicPartition("my_topic", 0, broker.BrokerID(), nil, nil, nil, vendor.ErrNoError)
	broker.Returns(metadataResponse)

	initProducerID := &vendor.InitProducerIDResponse{
		ThrottleTime:  0,
		ProducerID:    1000,
		ProducerEpoch: 1,
	}
	broker.Returns(initProducerID)

	config := vendor.NewConfig()
	config.Producer.Flush.Messages = 10
	config.Producer.Return.Successes = true
	config.Producer.Retry.Max = 4
	config.Producer.RequiredAcks = vendor.WaitForAll
	config.Producer.Retry.Backoff = 0
	config.Producer.Idempotent = true
	config.Net.MaxOpenRequests = 1
	config.Version = vendor.V0_11_0_0
	producer, err := vendor.NewAsyncProducer([]string{broker.Addr()}, config)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 10; i++ {
		producer.Input() <- &vendor.ProducerMessage{Topic: "my_topic", Key: nil, Value: vendor.StringEncoder(TestMessage)}
	}

	prodSuccess := &vendor.ProduceResponse{
		Version:      3,
		ThrottleTime: 0,
	}
	prodSuccess.AddTopicPartition("my_topic", 0, vendor.ErrNoError)
	broker.Returns(prodSuccess)
	expectResults(t, producer, 10, 0)

	broker.Close()
	closeProducer(t, producer)
}

func TestAsyncProducerIdempotentRetryCheckBatch(t *testing.T) {
	//Logger = log.New(os.Stderr, "", log.LstdFlags)
	tests := []struct {
		name           string
		failAfterWrite bool
	}{
		{"FailAfterWrite", true},
		{"FailBeforeWrite", false},
	}

	for _, test := range tests {
		broker := vendor.NewMockBroker(t, 1)

		metadataResponse := &vendor.MetadataResponse{
			Version:      1,
			ControllerID: 1,
		}
		metadataResponse.AddBroker(broker.Addr(), broker.BrokerID())
		metadataResponse.AddTopicPartition("my_topic", 0, broker.BrokerID(), nil, nil, nil, vendor.ErrNoError)

		initProducerIDResponse := &vendor.InitProducerIDResponse{
			ThrottleTime:  0,
			ProducerID:    1000,
			ProducerEpoch: 1,
		}

		prodNotLeaderResponse := &vendor.ProduceResponse{
			Version:      3,
			ThrottleTime: 0,
		}
		prodNotLeaderResponse.AddTopicPartition("my_topic", 0, vendor.ErrNotEnoughReplicas)

		prodDuplicate := &vendor.ProduceResponse{
			Version:      3,
			ThrottleTime: 0,
		}
		prodDuplicate.AddTopicPartition("my_topic", 0, vendor.ErrDuplicateSequenceNumber)

		prodOutOfSeq := &vendor.ProduceResponse{
			Version:      3,
			ThrottleTime: 0,
		}
		prodOutOfSeq.AddTopicPartition("my_topic", 0, vendor.ErrOutOfOrderSequenceNumber)

		prodSuccessResponse := &vendor.ProduceResponse{
			Version:      3,
			ThrottleTime: 0,
		}
		prodSuccessResponse.AddTopicPartition("my_topic", 0, vendor.ErrNoError)

		prodCounter := 0
		lastBatchFirstSeq := -1
		lastBatchSize := -1
		lastSequenceWrittenToDisk := -1
		handlerFailBeforeWrite := func(req *vendor.request) (res vendor.encoder) {
			switch req.body.key() {
			case 3:
				return metadataResponse
			case 22:
				return initProducerIDResponse
			case 0:
				prodCounter++

				preq := req.body.(*vendor.ProduceRequest)
				batch := preq.records["my_topic"][0].RecordBatch
				batchFirstSeq := int(batch.FirstSequence)
				batchSize := len(batch.Records)

				if lastSequenceWrittenToDisk == batchFirstSeq-1 { //in sequence append
					if lastBatchFirstSeq == batchFirstSeq { //is a batch retry
						if lastBatchSize == batchSize { //good retry
							// mock write to disk
							lastSequenceWrittenToDisk = batchFirstSeq + batchSize - 1
							return prodSuccessResponse
						}
						t.Errorf("[%s] Retried Batch firstSeq=%d with different size old=%d new=%d", test.name, batchFirstSeq, lastBatchSize, batchSize)
						return prodOutOfSeq
					} // not a retry
					// save batch just received for future check
					lastBatchFirstSeq = batchFirstSeq
					lastBatchSize = batchSize

					if prodCounter%2 == 1 {
						if test.failAfterWrite {
							// mock write to disk
							lastSequenceWrittenToDisk = batchFirstSeq + batchSize - 1
						}
						return prodNotLeaderResponse
					}
					// mock write to disk
					lastSequenceWrittenToDisk = batchFirstSeq + batchSize - 1
					return prodSuccessResponse
				}
				if lastBatchFirstSeq == batchFirstSeq && lastBatchSize == batchSize { // is a good batch retry
					if lastSequenceWrittenToDisk == (batchFirstSeq + batchSize - 1) { // we already have the messages
						return prodDuplicate
					}
					// mock write to disk
					lastSequenceWrittenToDisk = batchFirstSeq + batchSize - 1
					return prodSuccessResponse
				} //out of sequence / bad retried batch
				if lastBatchFirstSeq == batchFirstSeq && lastBatchSize != batchSize {
					t.Errorf("[%s] Retried Batch firstSeq=%d with different size old=%d new=%d", test.name, batchFirstSeq, lastBatchSize, batchSize)
				} else if lastSequenceWrittenToDisk+1 != batchFirstSeq {
					t.Errorf("[%s] Out of sequence message lastSequence=%d new batch starts at=%d", test.name, lastSequenceWrittenToDisk, batchFirstSeq)
				} else {
					t.Errorf("[%s] Unexpected error", test.name)
				}

				return prodOutOfSeq
			}
			return nil
		}

		config := vendor.NewConfig()
		config.Version = vendor.V0_11_0_0
		config.Producer.Idempotent = true
		config.Net.MaxOpenRequests = 1
		config.Producer.RequiredAcks = vendor.WaitForAll
		config.Producer.Return.Successes = true
		config.Producer.Flush.Frequency = 50 * time.Millisecond
		config.Producer.Retry.Backoff = 100 * time.Millisecond

		broker.setHandler(handlerFailBeforeWrite)
		producer, err := vendor.NewAsyncProducer([]string{broker.Addr()}, config)
		if err != nil {
			t.Fatal(err)
		}

		for i := 0; i < 3; i++ {
			producer.Input() <- &vendor.ProducerMessage{Topic: "my_topic", Key: nil, Value: vendor.StringEncoder(TestMessage)}
		}

		go func() {
			for i := 0; i < 7; i++ {
				producer.Input() <- &vendor.ProducerMessage{Topic: "my_topic", Key: nil, Value: vendor.StringEncoder("goroutine")}
				time.Sleep(100 * time.Millisecond)
			}
		}()

		expectResults(t, producer, 10, 0)

		broker.Close()
		closeProducer(t, producer)
	}
}

func TestAsyncProducerIdempotentErrorOnOutOfSeq(t *testing.T) {
	broker := vendor.NewMockBroker(t, 1)

	metadataResponse := &vendor.MetadataResponse{
		Version:      1,
		ControllerID: 1,
	}
	metadataResponse.AddBroker(broker.Addr(), broker.BrokerID())
	metadataResponse.AddTopicPartition("my_topic", 0, broker.BrokerID(), nil, nil, nil, vendor.ErrNoError)
	broker.Returns(metadataResponse)

	initProducerID := &vendor.InitProducerIDResponse{
		ThrottleTime:  0,
		ProducerID:    1000,
		ProducerEpoch: 1,
	}
	broker.Returns(initProducerID)

	config := vendor.NewConfig()
	config.Producer.Flush.Messages = 10
	config.Producer.Return.Successes = true
	config.Producer.Retry.Max = 400000
	config.Producer.RequiredAcks = vendor.WaitForAll
	config.Producer.Retry.Backoff = 0
	config.Producer.Idempotent = true
	config.Net.MaxOpenRequests = 1
	config.Version = vendor.V0_11_0_0

	producer, err := vendor.NewAsyncProducer([]string{broker.Addr()}, config)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 10; i++ {
		producer.Input() <- &vendor.ProducerMessage{Topic: "my_topic", Key: nil, Value: vendor.StringEncoder(TestMessage)}
	}

	prodOutOfSeq := &vendor.ProduceResponse{
		Version:      3,
		ThrottleTime: 0,
	}
	prodOutOfSeq.AddTopicPartition("my_topic", 0, vendor.ErrOutOfOrderSequenceNumber)
	broker.Returns(prodOutOfSeq)
	expectResults(t, producer, 0, 10)

	broker.Close()
	closeProducer(t, producer)
}

// This example shows how to use the producer while simultaneously
// reading the Errors channel to know about any failures.
func ExampleAsyncProducer_select() {
	producer, err := vendor.NewAsyncProducer([]string{"localhost:9092"}, nil)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := producer.Close(); err != nil {
			log.Fatalln(err)
		}
	}()

	// Trap SIGINT to trigger a shutdown.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	var enqueued, errors int
ProducerLoop:
	for {
		select {
		case producer.Input() <- &vendor.ProducerMessage{Topic: "my_topic", Key: nil, Value: vendor.StringEncoder("testing 123")}:
			enqueued++
		case err := <-producer.Errors():
			log.Println("Failed to produce message", err)
			errors++
		case <-signals:
			break ProducerLoop
		}
	}

	log.Printf("Enqueued: %d; errors: %d\n", enqueued, errors)
}

// This example shows how to use the producer with separate goroutines
// reading from the Successes and Errors channels. Note that in order
// for the Successes channel to be populated, you have to set
// config.Producer.Return.Successes to true.
func ExampleAsyncProducer_goroutines() {
	config := vendor.NewConfig()
	config.Producer.Return.Successes = true
	producer, err := vendor.NewAsyncProducer([]string{"localhost:9092"}, config)
	if err != nil {
		panic(err)
	}

	// Trap SIGINT to trigger a graceful shutdown.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	var (
		wg                          sync.WaitGroup
		enqueued, successes, errors int
	)

	wg.Add(1)
	go func() {
		defer wg.Done()
		for range producer.Successes() {
			successes++
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for err := range producer.Errors() {
			log.Println(err)
			errors++
		}
	}()

ProducerLoop:
	for {
		message := &vendor.ProducerMessage{Topic: "my_topic", Value: vendor.StringEncoder("testing 123")}
		select {
		case producer.Input() <- message:
			enqueued++

		case <-signals:
			producer.AsyncClose() // Trigger a shutdown of the producer.
			break ProducerLoop
		}
	}

	wg.Wait()

	log.Printf("Successfully produced: %d; errors: %d\n", successes, errors)
}
