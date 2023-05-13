# 线程池

一个基于Golang的协程池

### 如何使用
#### v1版本
> go get "github.com/Zhangzk6666/thread_pool/v1"

```go
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

```

#### 最新版本
> go get "github.com/Zhangzk6666/thread_pool"

```go
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

``