package main

import "fmt"

func main() {
	//加载配置
	filename := "conf/logagent.conf"
	err := LoadConf("ini",filename)
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
}
