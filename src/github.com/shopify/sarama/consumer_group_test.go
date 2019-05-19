package sarama

import (
	"context"
	"fmt"
	"vendor"
)

type exampleConsumerGroupHandler struct{}

func (exampleConsumerGroupHandler) Setup(_ vendor.ConsumerGroupSession) error   { return nil }
func (exampleConsumerGroupHandler) Cleanup(_ vendor.ConsumerGroupSession) error { return nil }
func (h exampleConsumerGroupHandler) ConsumeClaim(sess vendor.ConsumerGroupSession, claim vendor.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		fmt.Printf("Message topic:%q partition:%d offset:%d\n", msg.Topic, msg.Partition, msg.Offset)
		sess.MarkMessage(msg, "")
	}
	return nil
}

func ExampleConsumerGroup() {
	// Init config, specify appropriate version
	config := vendor.NewConfig()
	config.Version = vendor.V1_0_0_0
	config.Consumer.Return.Errors = true

	// Start with a client
	client, err := vendor.NewClient([]string{"localhost:9092"}, config)
	if err != nil {
		panic(err)
	}
	defer func() { _ = client.Close() }()

	// Start a new consumer group
	group, err := vendor.NewConsumerGroupFromClient("my-group", client)
	if err != nil {
		panic(err)
	}
	defer func() { _ = group.Close() }()

	// Track errors
	go func() {
		for err := range group.Errors() {
			fmt.Println("ERROR", err)
		}
	}()

	// Iterate over consumer sessions.
	ctx := context.Background()
	for {
		topics := []string{"my-topic"}
		handler := exampleConsumerGroupHandler{}

		err := group.Consume(ctx, topics, handler)
		if err != nil {
			panic(err)
		}
	}
}
