package main

import (
	"fmt"
	"github.com/shopify/sarama"
	"sync"
)
var topic = "nginx_log3"
var wg sync.WaitGroup
func main() {
	consumer, err := sarama.NewConsumer([]string{"localhost:9092"}, nil)
	if err != nil {
		fmt.Println("Failed to start consumer:%s",err)
		return
	}
	partitionList, err := consumer.Partitions(topic)
	if err != nil {
		fmt.Println("failed to get the list of partitions:",err)
		return
	}
	fmt.Println(partitionList)
	for partition := range partitionList {
		pc, err := consumer.ConsumePartition(topic, int32(partition), sarama.OffsetNewest)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer pc.AsyncClose()
		go func(sarama.PartitionConsumer) {
			wg.Add(1)
			for msg := range pc.Messages(){
				fmt.Printf("patition:%d, Offset:%d, Key:%s, Value:%s",msg.Partition,
					msg.Offset,msg.Key,msg.Value)
				fmt.Println()
			}
			wg.Done()
		}(pc)
		wg.Wait()
		consumer.Close()
	}

}
