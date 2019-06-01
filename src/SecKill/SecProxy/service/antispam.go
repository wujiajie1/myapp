package service

import (
	"fmt"
	"sync"
)

var (
	secLimitMgr = &SecLimitMgr{
		UserLimitMap:make(map[int]*SecLimit,10000),
	}
)
type SecLimitMgr struct {
	UserLimitMap map[int]*SecLimit
	lock sync.Mutex
}
type SecLimit struct {
	count int
	curTime int64
}
//检查是否超过限制，计数
//计算用户在1s内请求了多少次
func (p *SecLimit) Count(nowTime int64) int{
	//与当前请求时间进行比较、
	if p.curTime != nowTime {
		p.count = 1
		p.curTime = nowTime
		return p.count
	}
	p.count ++
	return p.count
}
//检查用户在1s内访问次数
func (p *SecLimit) Check(nowTime int64) int{
	if p.curTime != nowTime {
		return 0
	}
	return p.count
}
//请求频率检测
func antiSpam(req *SecRequest) error {
	secLimitMgr.lock.Lock()
	defer secLimitMgr.lock.Unlock()
	secLimit,ok := secLimitMgr.UserLimitMap[req.UserId]
	if !ok {
		secLimit = &SecLimit{}
		secLimitMgr.UserLimitMap[req.UserId] = secLimit
	}
	count := secLimit.Count(req.AccessTime.Unix())
	if count > secSkillConf.UserSecAccessLimit {
		err := fmt.Errorf("invalid request")
		return err
	}
	return nil
}