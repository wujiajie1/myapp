package sarama

import (
	"log"
	"sync"
	"testing"
	"vendor"
)

func TestSyncProducer(t *testing.T) {
	seedBroker := vendor.NewMockBroker(t, 1)
	leader := vendor.NewMockBroker(t, 2)

	metadataResponse := new(vendor.MetadataResponse)
	metadataResponse.AddBroker(leader.Addr(), leader.BrokerID())
	metadataResponse.AddTopicPartition("my_topic", 0, leader.BrokerID(), nil, nil, nil, vendor.ErrNoError)
	seedBroker.Returns(metadataResponse)

	prodSuccess := new(vendor.ProduceResponse)
	prodSuccess.AddTopicPartition("my_topic", 0, vendor.ErrNoError)
	for i := 0; i < 10; i++ {
		leader.Returns(prodSuccess)
	}

	producer, err := vendor.NewSyncProducer([]string{seedBroker.Addr()}, nil)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 10; i++ {
		msg := &vendor.ProducerMessage{
			Topic:    "my_topic",
			Value:    vendor.StringEncoder(vendor.TestMessage),
			Metadata: "test",
		}

		partition, offset, err := producer.SendMessage(msg)

		if partition != 0 || msg.Partition != partition {
			t.Error("Unexpected partition")
		}
		if offset != 0 || msg.Offset != offset {
			t.Error("Unexpected offset")
		}
		if str, ok := msg.Metadata.(string); !ok || str != "test" {
			t.Error("Unexpected metadata")
		}
		if err != nil {
			t.Error(err)
		}
	}

	vendor.safeClose(t, producer)
	leader.Close()
	seedBroker.Close()
}

func TestSyncProducerBatch(t *testing.T) {
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
	config.Producer.Flush.Messages = 3
	config.Producer.Return.Successes = true
	producer, err := vendor.NewSyncProducer([]string{seedBroker.Addr()}, config)
	if err != nil {
		t.Fatal(err)
	}

	err = producer.SendMessages([]*vendor.ProducerMessage{
		{
			Topic:    "my_topic",
			Value:    vendor.StringEncoder(vendor.TestMessage),
			Metadata: "test",
		},
		{
			Topic:    "my_topic",
			Value:    vendor.StringEncoder(vendor.TestMessage),
			Metadata: "test",
		},
		{
			Topic:    "my_topic",
			Value:    vendor.StringEncoder(vendor.TestMessage),
			Metadata: "test",
		},
	})

	if err != nil {
		t.Error(err)
	}

	vendor.safeClose(t, producer)
	leader.Close()
	seedBroker.Close()
}

func TestConcurrentSyncProducer(t *testing.T) {
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
	config.Producer.Flush.Messages = 100
	config.Producer.Return.Successes = true
	producer, err := vendor.NewSyncProducer([]string{seedBroker.Addr()}, config)
	if err != nil {
		t.Fatal(err)
	}

	wg := sync.WaitGroup{}

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			msg := &vendor.ProducerMessage{Topic: "my_topic", Value: vendor.StringEncoder(vendor.TestMessage)}
			partition, _, err := producer.SendMessage(msg)
			if partition != 0 {
				t.Error("Unexpected partition")
			}
			if err != nil {
				t.Error(err)
			}
			wg.Done()
		}()
	}
	wg.Wait()

	vendor.safeClose(t, producer)
	leader.Close()
	seedBroker.Close()
}

func TestSyncProducerToNonExistingTopic(t *testing.T) {
	broker := vendor.NewMockBroker(t, 1)

	metadataResponse := new(vendor.MetadataResponse)
	metadataResponse.AddBroker(broker.Addr(), broker.BrokerID())
	metadataResponse.AddTopicPartition("my_topic", 0, broker.BrokerID(), nil, nil, nil, vendor.ErrNoError)
	broker.Returns(metadataResponse)

	config := vendor.NewConfig()
	config.Metadata.Retry.Max = 0
	config.Producer.Retry.Max = 0
	config.Producer.Return.Successes = true

	producer, err := vendor.NewSyncProducer([]string{broker.Addr()}, config)
	if err != nil {
		t.Fatal(err)
	}

	metadataResponse = new(vendor.MetadataResponse)
	metadataResponse.AddTopic("unknown", vendor.ErrUnknownTopicOrPartition)
	broker.Returns(metadataResponse)

	_, _, err = producer.SendMessage(&vendor.ProducerMessage{Topic: "unknown"})
	if err != vendor.ErrUnknownTopicOrPartition {
		t.Error("Uxpected ErrUnknownTopicOrPartition, found:", err)
	}

	vendor.safeClose(t, producer)
	broker.Close()
}

func TestSyncProducerRecoveryWithRetriesDisabled(t *testing.T) {
	seedBroker := vendor.NewMockBroker(t, 1)
	leader1 := vendor.NewMockBroker(t, 2)
	leader2 := vendor.NewMockBroker(t, 3)

	metadataLeader1 := new(vendor.MetadataResponse)
	metadataLeader1.AddBroker(leader1.Addr(), leader1.BrokerID())
	metadataLeader1.AddTopicPartition("my_topic", 0, leader1.BrokerID(), nil, nil, nil, vendor.ErrNoError)
	seedBroker.Returns(metadataLeader1)

	config := vendor.NewConfig()
	config.Producer.Retry.Max = 0 // disable!
	config.Producer.Retry.Backoff = 0
	config.Producer.Return.Successes = true
	producer, err := vendor.NewSyncProducer([]string{seedBroker.Addr()}, config)
	if err != nil {
		t.Fatal(err)
	}
	seedBroker.Close()

	prodNotLeader := new(vendor.ProduceResponse)
	prodNotLeader.AddTopicPartition("my_topic", 0, vendor.ErrNotLeaderForPartition)
	leader1.Returns(prodNotLeader)
	_, _, err = producer.SendMessage(&vendor.ProducerMessage{Topic: "my_topic", Value: vendor.StringEncoder(vendor.TestMessage)})
	if err != vendor.ErrNotLeaderForPartition {
		t.Fatal(err)
	}

	metadataLeader2 := new(vendor.MetadataResponse)
	metadataLeader2.AddBroker(leader2.Addr(), leader2.BrokerID())
	metadataLeader2.AddTopicPartition("my_topic", 0, leader2.BrokerID(), nil, nil, nil, vendor.ErrNoError)
	leader1.Returns(metadataLeader2)
	prodSuccess := new(vendor.ProduceResponse)
	prodSuccess.AddTopicPartition("my_topic", 0, vendor.ErrNoError)
	leader2.Returns(prodSuccess)
	_, _, err = producer.SendMessage(&vendor.ProducerMessage{Topic: "my_topic", Value: vendor.StringEncoder(vendor.TestMessage)})
	if err != nil {
		t.Fatal(err)
	}

	leader1.Close()
	leader2.Close()
	vendor.safeClose(t, producer)
}

// This example shows the basic usage pattern of the SyncProducer.
func ExampleSyncProducer() {
	producer, err := vendor.NewSyncProducer([]string{"localhost:9092"}, nil)
	if err != nil {
		log.Fatalln(err)
	}
	defer func() {
		if err := producer.Close(); err != nil {
			log.Fatalln(err)
		}
	}()

	msg := &vendor.ProducerMessage{Topic: "my_topic", Value: vendor.StringEncoder("testing 123")}
	partition, offset, err := producer.SendMessage(msg)
	if err != nil {
		log.Printf("FAILED to send message: %s\n", err)
	} else {
		log.Printf("> message sent to partition %d at offset %d\n", partition, offset)
	}
}
