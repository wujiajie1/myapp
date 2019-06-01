package main

import (
	"github.com/astaxie/beego"
	_ "SecKill/SecProxy/router"
)

func main() {
	//加载配置 local ip:192.168.31.95
	err := InitConfig()
	if err != nil {
		panic(err)
		return
	}
	err = InitSec()
	if err != nil {
		panic(err)
		return
	}
	beego.Run()
}


