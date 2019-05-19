package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"syscall"
)

func main() {
	fileName := "conf/logagent.conf"
	dir, _ := os.Getwd()
	wd, _ := syscall.Getwd()
	fmt.Println(dir,wd)
	file, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(bytes))
}
