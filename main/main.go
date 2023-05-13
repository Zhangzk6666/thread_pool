package main

import (
	"fmt"
	"math/rand"
	"time"

	zk_pool "github.com/Zhangzk6666/thread_pool"
)

func main() {
	// pool := zk_pool.NewPool()
	pool := zk_pool.NewPoolWithSetParameter(20, 1050*time.Millisecond)
	go pool.Start()
	time.Sleep(1 * time.Second)
	for i := 0; i < 1000; i++ {
		pool.AddJob(func() {
			rand.Seed(time.Now().UnixNano())
			fmt.Println("job -------", rand.Intn(100000))
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
		})
	}
	// time.Sleep(1050 * time.Millisecond)
	// pool.Close <- struct{}{}
	time.Sleep(6 * time.Second)
}
