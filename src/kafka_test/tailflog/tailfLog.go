package main

import (
	"fmt"
	"time"
	"vendor"
)

func main() {
	fileName := "D:\\mysoftwore\\kafka_2.12-2.2.0\\logs\\controller.log.2019-05-15-08"
	tails, err := vendor.TailFile(fileName, vendor.Config{
		//Location: &tail.SeekInfo{Offset: 0, Whence: 2},
		ReOpen:   true,
		Poll:     true,
		Follow:   true,
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	var msg *vendor.Line
	var ok bool
	for true {
		msg,ok = <- tails.Lines
		if !ok {
			fmt.Printf("tail file close reopen, filename:%s.",fileName)
			time.Sleep(100*time.Millisecond)
			continue
		}
		fmt.Printf("msg:%s\n",msg.Text)
	}

}
