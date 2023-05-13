package main

import (
	"fmt"
	"math/rand"
	zk_pool "github.com/Zhangzk6666/thread_pool/v1"
	"time"
)

func main() {
	pool := zk_pool.NewPool(5, 2*time.Second)
	go pool.Start()
	time.Sleep(1 * time.Second)
	for i := 0; i < 100; i++ {
		pool.AddJob(func() {
			fmt.Println("job -------")
			rand.Seed(time.Now().UnixNano())
			time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
		})
	}
	// time.Sleep(1050 * time.Millisecond)
	// pool.Close <- struct{}{}
	time.Sleep(3 * time.Second)
}
