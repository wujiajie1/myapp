package main

import (
	"SecKill/SecProxy/service"
	"fmt"
	"github.com/astaxie/beego"
	"strings"
)

var (
	secKillConf = &service.SecSkillConf{
		SecProductInfoMap:make(map[int]*service.SecProductInfoConf,1024),
	}
)

func InitConfig() error{
	redisAddr := beego.AppConfig.String("redis_addr")
	etcdAddr := beego.AppConfig.String("etcd_addr")
	secKillConf.EtcdConf.EtcdAddr = etcdAddr
	secKillConf.RedisConf.RedisAddr = redisAddr
	if len(redisAddr) == 0 || len(etcdAddr) == 0 {
		err := fmt.Errorf("init config failed,redis[%s] or etcd[%s] config is null",redisAddr,etcdAddr)
		return err
	}
	redisMaxIdle, err := beego.AppConfig.Int("redis_black_idle")
	if err != nil {
		err = fmt.Errorf("init config failed, read redis_black_idle error:%v", err)
		return err
	}

	redisMaxActive, err := beego.AppConfig.Int("redis_black_active")
	if err != nil {
		err = fmt.Errorf("init config failed, read redis_black_active error:%v", err)
		return err
	}

	redisIdleTimeout, err := beego.AppConfig.Int("redis_black_idle_timeout")
	if err != nil {
		err = fmt.Errorf("init config failed, read redis_black_idle_timeout error:%v", err)
		return err
	}

	secKillConf.RedisConf.RedisMaxIdle = redisMaxIdle
	secKillConf.RedisConf.RedisMaxActive = redisMaxActive
	secKillConf.RedisConf.RedisIdleTimeout = redisIdleTimeout

	etcdTimeout, err := beego.AppConfig.Int("etcd_timeout")
	if err != nil {
		err = fmt.Errorf("init config failed, read etcd_timeout error:%v", err)
		return err
	}

	secKillConf.EtcdConf.Timeout = etcdTimeout
	secKillConf.EtcdConf.EtcdSecKeyPrefix = beego.AppConfig.String("etcd_sec_key_prefix")
	if len(secKillConf.EtcdConf.EtcdSecKeyPrefix) == 0 {
		err = fmt.Errorf("init config failed,read etcd_sec_key_prefix failed,err:%v",err)
		return err
	}
	productKey := beego.AppConfig.String("etcd_product_key")
	if len(productKey) == 0 {
		err = fmt.Errorf("init config failed, read etcd_product_key error:%v", err)
		return err
	}
	if strings.HasSuffix(secKillConf.EtcdConf.EtcdSecKeyPrefix, "/") == false {
		secKillConf.EtcdConf.EtcdSecKeyPrefix = secKillConf.EtcdConf.EtcdSecKeyPrefix + "/"
	}
	secKillConf.EtcdConf.EtcdSecProductKey = fmt.Sprintf("%s%s", secKillConf.EtcdConf.EtcdSecKeyPrefix, productKey)

	secKillConf.LogPath = beego.AppConfig.String("log_path")
	secKillConf.LogLevel = beego.AppConfig.String("log_level")

	secKillConf.CookieSecretKey = beego.AppConfig.String("cookie_secretkey")
	secKillConf.UserSecAccessLimit,_ = beego.AppConfig.Int("user_sec_access_limit")

	return nil
}
