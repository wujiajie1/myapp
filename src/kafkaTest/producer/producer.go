package main

import (
	"fmt"
	"github.com/shopify/sarama"
	"log"
	"os"
	"time"
)

var Addr = []string{"localhost:9092"}
func main() {
	syncProducer()
	//SaramaProducer()
}

func syncProducer() {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Timeout = 5 * time.Second

	p, err := sarama.NewSyncProducer(Addr, config)
	if err != nil {
		log.Printf("sarama.NewSyncProducer err, message=%s \n", err)
		return
	}
	defer p.Close()
	topic := "test"
	srcValue := "sync: this is a message. index=%d"
	for i:=0; i<10; i++ {
		value := fmt.Sprintf(srcValue, i)
		msg := &sarama.ProducerMessage{
			Topic:topic,
			Value:sarama.ByteEncoder(value),
		}
		part, offset, err := p.SendMessage(msg)
		if err != nil {
			log.Printf("send message(%s) err=%s \n", value, err)
		}else {
			fmt.Fprintf(os.Stdout, value + "发送成功，partition=%d, offset=%d \n", part, offset)
		}
		time.Sleep(2*time.Second)
	}
}

func SaramaProducer()  {

	config := sarama.NewConfig()
	//等待服务器所有副本都保存成功后的响应
	config.Producer.RequiredAcks = sarama.WaitForAll
	//随机向partition发送消息
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	//设置使用的kafka版本,如果低于V0_10_0_0版本,消息中的timestrap没有作用.需要消费和生产同时配置
	//注意，版本设置不对的话，kafka会返回很奇怪的错误，并且无法成功发送消息
	config.Version = sarama.V2_1_0_0
	fmt.Println("start make producer")
	//使用配置,新建一个异步生产者
	producer, e := sarama.NewAsyncProducer(Addr, config)
	if e != nil {
		fmt.Println(e)
		return
	}
	defer producer.AsyncClose()

	fmt.Println("start goroutine")
	go func(p sarama.AsyncProducer) {
		for{
			select {
			case  <-p.Successes():
				//fmt.Println("offset: ", suc.Offset, "timestamp: ", suc.Timestamp.String(), "partitions: ", suc.Partition)
			case fail := <-p.Errors():
				fmt.Println("err: ", fail.Err)
			}
		}
	}(producer)

	var value string
	for i:=0;;i++ {
		time.Sleep(500 * time.Millisecond)
		time11 := time.Now()
		value = "this is a message 0606 " + time11.Format("15:04:05")

		// 发送的消息,主题。
		// 注意：这里的msg必须得是新构建的变量，不然你会发现发送过去的消息内容都是一样的，因为批次发送消息的关系。
		msg := &sarama.ProducerMessage{
			Topic: "0606_test",
		}

		//将字符串转化为字节数组
		msg.Value = sarama.ByteEncoder(value)
		//fmt.Println(value)

		//使用通道发送
		producer.Input() <- msg
	}
}
