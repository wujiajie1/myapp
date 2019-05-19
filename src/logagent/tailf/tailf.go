package tailf

import (
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/hpcloud/tail"
	"logagent/module"
	"time"
)
type TailObj struct {
	tail *tail.Tail
	conf module.CollectConf
}
type TextMsg struct {
	Msg string
	Topic string
}
type TailObjMgr struct {
	tailObjs []*TailObj
	msgChan chan *TextMsg
}

var (
	tailObjMgr *TailObjMgr
)
func InitTail(config *module.Config) error {
	fmt.Println(config.ChanSize)
	if len(config.Collect) == 0{
		err := fmt.Errorf("invalid config for log collect,conf:%v",config.Collect)
		return err
	}
	tailObjMgr = &TailObjMgr{
		msgChan: make(chan *TextMsg,config.ChanSize),
	}
	for _,v := range config.Collect{
		tails, err := tail.TailFile(v.LogPath, tail.Config{
			ReOpen:    true,
			Follow:    true,
			MustExist: false,
			Poll:      true,
		})
		if err != nil {
			fmt.Println("tail file err,err:%v",err)
			return err
		}
		obj := &TailObj{
			conf:v,
			tail:tails,
		}
		tailObjMgr.tailObjs = append(tailObjMgr.tailObjs,obj)
		go readFromTail(obj)
	}

	return nil
}

func readFromTail(tailObj *TailObj) {
	for true {
		msg,ok := <- tailObj.tail.Lines
		if !ok {
			logs.Warn("tail file close reopen,filename:%s\n",tailObj.tail.Filename)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		fmt.Println("msg:",msg.Text)
		textMsg := &TextMsg{
			Msg: msg.Text,
			Topic: tailObj.conf.Topic,
		}
		fmt.Println(textMsg)
		tailObjMgr.msgChan <-textMsg
	}
}

func GetOneLine()(msg *TextMsg){
	msgdata := <- tailObjMgr.msgChan
	fmt.Println("getmsg:",msgdata)
	return msgdata
}
