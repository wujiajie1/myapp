package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

func main() {
	process()
}
type Result struct {
	r   *http.Response
	err error
}
func process() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	tr := &http.Transport{}
	//client := http.Client{Transport: tr}
	c := make(chan Result, 1)
	req, err := http.NewRequest("GET", "Http://www.google.com", nil)
	if err != nil {
		fmt.Println("http request failed, err:", err)
		return
	}
	go func() {
		//resp, err := client.Do(req)
		//pack := Result{r: resp, err: err}
		//c <- pack
		time.Sleep(3*time.Second)
	}()
	select {
	case <- ctx.Done():
		tr.CancelRequest(req)
		//res := <- c
		fmt.Println("Timeout! err:")
	case res := <-c:
		defer res.r.Body.Close()
		out, _ := ioutil.ReadAll(res.r.Body)
		fmt.Printf("Server Response: %s", out)
	}
	return
}
