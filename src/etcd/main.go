package main

import (
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"golang.org/x/net/context"
	"time"
)

func EtcdExample(){
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		fmt.Println("connect faild,err:",err)
		return
	}
	fmt.Println("connect succ")
	defer cli.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	_, err = cli.Put(ctx, "/logagent/main/conf/", "sample_value")
	cancel()
	if err != nil {
		fmt.Println("put failed,err:",err)
	}
	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	resp, err := cli.Get(ctx, "/logagent/main/conf/")
	if err != nil {
		fmt.Println("get failed,err:",err)
		return
	}
	for _,ev := range resp.Kvs {
		fmt.Printf("%s : %s \n",ev.Key,ev.Value)
	}
}
func main() {
	SetLogConfToEtcd()
}

func SetLogConfToEtcd() {

}
