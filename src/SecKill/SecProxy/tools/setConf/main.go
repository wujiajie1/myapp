package main

import (
	"encoding/json"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"time"
	"golang.org/x/net/context"
)

const (
	EtcdKey = "/oldboy/backend/secskill/product"
)
type SecInfoConf struct {
	ProductId int
	StartTime int
	EndTime int
	Status  int
	Count   int
	Left    int   //商品剩余量
}
func main() {
	SetLogConfToEtcd()
}

func SetLogConfToEtcd() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379", "localhost:22379", "localhost:32379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		fmt.Println("connect failed, err:", err)
		return
	}

	fmt.Println("connect succ")
	defer cli.Close()
	var secInfoArr []SecInfoConf
	secInfoArr = append(secInfoArr,SecInfoConf{
		1028,
		1559313298,
		15593832933,
		0,
		1000,
		1000,
	})
	secInfoArr = append(secInfoArr,SecInfoConf{
		1027,
		1559313298,
		15593832933,
		0,
		2000,
		2000,
	})
	secInfoArr = append(secInfoArr,SecInfoConf{
		1026,
		1559313298,
		15593832933,
		0,
		2000,
		2000,
	})
	data, err := json.Marshal(secInfoArr)
	if err != nil {
		fmt.Println("json failed, ", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	//cli.Delete(ctx, EtcdKey)
	//return
	_, err = cli.Put(ctx, EtcdKey, string(data))
	defer cancel()
	if err != nil {
		fmt.Println("put failed, err:", err)
		return
	}
	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	resp, err := cli.Get(ctx, EtcdKey)
	defer cancel()
	if err != nil {
		fmt.Println("get failed, err:", err)
		return
	}
	for _, ev := range resp.Kvs {
		fmt.Printf("%s : %s\n", ev.Key, ev.Value)
	}
}
