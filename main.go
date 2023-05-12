package main

import (
	"fmt"
	"math/rand"
	"sync"
	zk_pool "thread_pool/pool"
	"time"
)

func main() {
	pool := &zk_pool.Pool{
		Size:    10,
		Cache:   make(chan func()),
		Workers: []*zk_pool.Worker{},
		Close:   make(chan struct{}),
		Timeout: 2 * time.Second,
	}
	for i := 0; i < pool.Size; i++ {
		pool.Workers = append(pool.Workers, &zk_pool.Worker{
			Id:                     i,
			IsWork:                 false,
			Job:                    make(chan func()),
			Shutdown:               make(chan struct{}),
			Mutex:                  sync.Mutex{},
			CumulativeNumberOfJobs: 0,
		})
	}
	go pool.Start()
	time.Sleep(1 * time.Second)
	for i := 0; i < 1000; i++ {
		pool.Cache <- func() {
			fmt.Println("job -------")
			rand.Seed(time.Now().UnixNano())
			time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
		}
	}
	// time.Sleep(1050 * time.Millisecond)
	// pool.Close <- struct{}{}
	time.Sleep(3 * time.Second)
}
