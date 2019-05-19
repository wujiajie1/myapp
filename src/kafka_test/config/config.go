package main

import (
	"fmt"
	"vendor"
)

func main() {
	fileName := "D:\\workspace\\golang\\myapp\\src\\kafka_test\\config\\test.ini"
	conf, err := vendor.NewConfig("ini", fileName)
	if err != nil {
		panic(err)
	}
	i, err := conf.Int("port")
	if err != nil {
		panic(err)
	}
	fmt.Println(i)
}
