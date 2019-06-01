package main

import (
	"SecKill/SecProxy/service"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/garyburd/redigo/redis"
	etcd_client "github.com/coreos/etcd/clientv3"
	"golang.org/x/net/context"
	"time"
)
var (
	redisPool  *redis.Pool
	etcdClient *etcd_client.Client
)
//初始化秒杀应用
func InitSec() error {
	err := initLogs()
	if err != nil {
		logs.Error("init logger failed,err:%v",err)
		return err
	}

	err = initRedis()
	if err != nil {
		logs.Error("init redis failed,err:%v",err)
		return err
	}

	err = initEtcd()
	if err != nil {
		logs.Error("init etcd failed,err:%v",err)
		return err
	}
	err = loadSecConf()
	if err != nil {
		logs.Error("load Sec Conf failed,err:%v",err)
		return err
	}
	service.InitService(secKillConf)
	initSecProductWatcher()

	logs.Info("init sec succ")
	return nil
}

func initSecProductWatcher() {
	go watchSecProductKey(secKillConf.EtcdConf.EtcdSecProductKey)
}

func convertLogLevel(level string) int {

	switch level {
	case "debug":
		return logs.LevelDebug
	case "warn":
		return logs.LevelWarn
	case "info":
		return logs.LevelInfo
	case "trace":
		return logs.LevelTrace
	}

	return logs.LevelDebug
}
func initLogs() error {
	config := make(map[string]interface{})
	config["filename"] = secKillConf.LogPath
	config["level"] = convertLogLevel(secKillConf.LogLevel)

	configStr, err := json.Marshal(config)
	if err != nil {
		fmt.Println("marshal failed, err:", err)
		return err
	}

	logs.SetLogger(logs.AdapterFile, string(configStr))

	return nil
}

func initRedis() error {
	redisPool = &redis.Pool{
		MaxIdle: secKillConf.RedisConf.RedisMaxIdle,//空闲连接数
		MaxActive: secKillConf.RedisConf.RedisMaxActive, //活跃连接数,0表示无限制
		IdleTimeout: time.Duration(secKillConf.RedisConf.RedisIdleTimeout), //空闲超时时间
		Dial: func() (conn redis.Conn, e error) {
			return redis.Dial("tcp",secKillConf.RedisConf.RedisAddr)
		},
	}
	conn := redisPool.Get()
	defer conn.Close()
	_, err := conn.Do("ping")
	if err != nil {
		logs.Error("ping redis failed,err:%v",err)
		return err
	}
	return nil
}

func initEtcd() error {
	cli, err := etcd_client.New(etcd_client.Config{
		Endpoints:   []string{secKillConf.EtcdConf.EtcdAddr},
		DialTimeout: time.Duration(secKillConf.EtcdConf.Timeout) * time.Second,
	})
	if err != nil {
		logs.Error("connect etcd failed, err:", err)
		return err
	}

	etcdClient = cli
	return nil
}

func loadSecConf() error {
	//从etcd获取配置
	//预先定义一个key,从配置文件中获取
	resp, err := etcdClient.Get(context.Background(), secKillConf.EtcdConf.EtcdSecProductKey)
	if err != nil{
		logs.Error("get [%s] from etcd failed,err:%v",secKillConf.EtcdConf.EtcdSecProductKey,err)
		return err
	}
	var secProductInfo []service.SecProductInfoConf
	for k, v := range resp.Kvs {
		logs.Debug("key[%v] valud[%v]", k, v)
		err = json.Unmarshal(v.Value, &secProductInfo)
		if err != nil {
			logs.Error("Unmarshal sec product info failed, err:%v", err)
			return err
		}

		logs.Debug("sec info conf is [%v]", secProductInfo)
	}
	updateSecProductInfo(secProductInfo)

	return nil
}
func watchSecProductKey(key string) {
	cli, err := etcd_client.New(etcd_client.Config{
		Endpoints:   []string{"localhost:2379", "localhost:22379", "localhost:32379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		logs.Error("connect etcd failed, err:", err)
		return
	}

	logs.Debug("begin watch key:%s", key)
	for {
		rch := cli.Watch(context.Background(), key)
		var secProductInfo []service.SecProductInfoConf
		var getConfSucc = true

		for wresp := range rch {
			for _, ev := range wresp.Events {
				if ev.Type == mvccpb.DELETE {
					logs.Warn("key[%s] 's config deleted", key)
					continue
				}

				if ev.Type == mvccpb.PUT && string(ev.Kv.Key) == key {
					err = json.Unmarshal(ev.Kv.Value, &secProductInfo)
					if err != nil {
						logs.Error("key [%s], Unmarshal[%s], err:%v ", err)
						getConfSucc = false
						continue
					}
				}
				logs.Debug("get config from etcd, %s %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
			}

			if getConfSucc {
				logs.Debug("get config from etcd succ, %v", secProductInfo)
				updateSecProductInfo(secProductInfo)
			}
		}

	}
}

func updateSecProductInfo(secProductInfo []service.SecProductInfoConf) {
	tmp := make(map[int]*service.SecProductInfoConf,1024)

	for _,v := range secProductInfo{
		product := v
		tmp[v.ProductId] = &product
	}
	secKillConf.RwSecProductLock.Lock()
	secKillConf.SecProductInfoMap = tmp
	secKillConf.RwSecProductLock.Unlock()
}