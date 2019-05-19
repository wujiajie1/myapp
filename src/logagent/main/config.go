package main

import (
	"errors"
	"fmt"
	"github.com/astaxie/beego/config"
)
var (
	appConfig *Config
)
//config 存取加载的配置
type Config struct {
	LogLevel 	string `json:"log_level"`
	LogPath 	string `json:"log_path"`
	Collect 	[]CollectConf `json:"collect"`
}
//CollectConf 日志收集配置
type CollectConf struct {
	LogPath 	string `json:"log_path"`
	Topic 		string `json:"topic"`
}
func LoadConf(confType, fileName string) error {
	conf, err := config.NewConfig(confType, fileName)
	if err != nil {
		fmt.Println("new config failed,err:",err)
		return err
	}
	//生成config的实例
	appConfig = &Config{}
	appConfig.LogLevel = conf.String("logs::log_level")
	//容错处理
	if len(appConfig.LogLevel) == 0 {
		appConfig.LogLevel = "debug"
	}
	appConfig.LogPath = conf.String("logs::log_path")
	//容错处理
	if len(appConfig.LogPath) == 0 {
		appConfig.LogPath = "./logs/logagent.log"
	}
	err = LoadCollectConf(conf)
	if err != nil {
		fmt.Println("load collect conf failed,err:",err)
		return err
	}
	return nil
}

func LoadCollectConf(configer config.Configer) error {
	var cc CollectConf
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
