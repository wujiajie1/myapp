package main

import (
	"fmt"
	"github.com/shopify/sarama"
)

func main() {
	consumer, err := sarama.NewConsumer([]string{"localhost:9092"}, nil)
	if err != nil {
		fmt.Println("Failed to start consumer:%s",err)
		return
	}
	partitionList, err := consumer.Partitions("nginx_log")
	if err != nil {
		fmt.Println("failed to get the list of partitions:",err)
		return
	}
	fmt.Println(partitionList)
}
