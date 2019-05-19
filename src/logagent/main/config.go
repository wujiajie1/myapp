package main

import (
	"errors"
	"fmt"
	"github.com/astaxie/beego/config"
	"logagent/module"
)
var (
	appConfig *module.Config
)

func LoadConf(confType, fileName string) (*module.Config,error) {
	conf, err := config.NewConfig(confType, fileName)
	if err != nil {
		fmt.Println("new config failed,err:",err)
		return nil,err
	}
	//生成config的实例
	appConfig = &module.Config{}
	appConfig.LogLevel = conf.String("logs::log_level")
	if len(appConfig.LogLevel) == 0 {
		appConfig.LogLevel = "debug"
	}

	appConfig.LogPath = conf.String("logs::log_path")
	if len(appConfig.LogPath) == 0 {
		appConfig.LogPath = "./logs/logagent.log"
	}

	appConfig.KafkaAddr = conf.String("logs::kafka_addr")
	if len(appConfig.KafkaAddr) == 0 {
		appConfig.KafkaAddr = "localhost:9092"
	}

	appConfig.ChanSize, err = conf.Int("logs::chan_size")
	if err != nil {
		fmt.Println("load chan_size conf failed,err:",err)
		appConfig.ChanSize = 100
	}
	err = LoadCollectConf(conf)
	if err != nil {
		fmt.Println("load collect conf failed,err:",err)
		return nil,err
	}
	return appConfig,nil
}

func LoadCollectConf(configer config.Configer) error {
	var cc module.CollectConf
	var err error
	cc.LogPath = configer.String("collect::log_path")
	if len(cc.LogPath) == 0 {
		err = errors.New("invalid collect::log_path")
		return err
	}

	cc.Topic = configer.String("collect::topic")
	if len(cc.LogPath) == 0 {
		err = errors.New("invalid collect::topic")
		return err
	}

	appConfig.Collect = append(appConfig.Collect,cc)
	return err
}
