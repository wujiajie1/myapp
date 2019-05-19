package main

import (
	"fmt"
	"log"
	"os"
	"time"
	"vendor"
)

var Address = []string{"127.0.0.1:9092"}
func main() {
	syncProducer()
}
func syncProducer(){
	config := vendor.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Timeout = 5*time.Second
	producer, err := vendor.NewSyncProducer(Address, config)
	if err != nil {
		log.Printf("sarama.NewSyncProducer err,message=%s \n",err.Error())
		return
	}
	defer producer.Close()
	topic := "test3"
	srcValue := "sync: this is a message. index=%d"
	for i:=0; i < 10;i++{
		value := fmt.Sprintf(srcValue, i)
		fmt.Println(value)
		msg := &vendor.ProducerMessage{
			Topic: topic,
			Value: vendor.ByteEncoder(value),
		}
		partition, offset, err := producer.SendMessage(msg)
		if err != nil {
			log.Printf("send message(%s) err=%s \n", value, err)
		}else {
			fmt.Fprintf(os.Stdout, value + "发送成功，partition=%d, offset=%d \n", partition, offset)
		}
		time.Sleep(2*time.Second)
	}
}
