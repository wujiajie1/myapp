package main

import (
	"fmt"
	"github.com/astaxie/beego/logs"
	"logagent/kafka"
	"logagent/tailf"
	"os"
)

func main() {
	//加载配置
	sysdir, err := os.Getwd()
	if err != nil {
		fmt.Printf("get sys path failed,err:%v\n",err)
		panic("get sys path failed")
		return
	}
	filename := sysdir+"/conf/logagent.conf"
	appConfig, err := LoadConf("ini", filename)
	if err != nil {
		fmt.Printf("load conf failed,err:%v\n",err)
		panic("load conf failed")
		return
	}
	//初始化日志
	err = initLogger()
	if err != nil {
		fmt.Printf("load logger failed, err:%v\n", err)
		panic("load logger failed")
		return
	}
	logs.Debug("init succ")
	logs.Debug("log conf succ,config:%v",appConfig)

	err = tailf.InitTail(appConfig)
	if err != nil {
		logs.Error("init tail failed,err:%v",err)
		return
	}
	logs.Debug("init tailf succ")
	err = kafka.InitKafka(appConfig.KafkaAddr)
	if err != nil {
		logs.Error("init kafka failed,err:%v",err)
		return
	}
	logs.Debug("init kafka succ")
	err = serverRun()
	if err != nil {
		logs.Error("serverRun failed,err:%v",err)
		return
	}
	logs.Info("program exited")
}
