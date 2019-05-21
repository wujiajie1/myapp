package main

import (
	"context"
	"fmt"
)

func main() {
	ctx := context.WithValue(context.Background(), "trace_id", 111111)
	process(ctx)
}

func process(ctx context.Context) {
	ret,ok := ctx.Value("trace_id").(int)
	if !ok {
		return
	}
	b := ret + 2
	fmt.Printf("type:%T,value:%d",ret,b)
}
