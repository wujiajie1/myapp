package main

import (
	"github.com/astaxie/beego/logs"
	"logagent/kafka"
	"logagent/tailf"
	"time"
)

func serverRun() error {

	for{
		msg := tailf.GetOneLine()
		err := SendTokafka(msg)
		if err != nil {
			logs.Error("send to kafka failed,err:%v",err)
			time.Sleep(time.Second)
			continue
		}
	}
	return nil
}

func SendTokafka(msg *tailf.TextMsg) error {
	//logs.Debug("read msg:%s,topic:%s",msg.Msg,msg.Topic)
	return kafka.SendToKafka(msg.Msg,msg.Topic)
}
