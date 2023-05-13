package pool

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

type Pool struct {
	size     int           // 协程个数
	jobCache chan func()   // 等待执行的函数
	workers  []*worker     // 协程集合（包括正在工作的和空闲的，也就是全部）
	close    chan struct{} // 关闭协程池的信号
	timeout  time.Duration // 超时关闭线程池
}

// 创建线程池
func NewPool(size int, timeout time.Duration) *Pool {
	if size <= 0 {
		size = runtime.NumCPU()
	}
	return &Pool{
		size:     size,
		jobCache: make(chan func()),
		workers:  []*worker{},
		close:    make(chan struct{}),
		timeout:  timeout,
	}
}

// 启动协程池
func (p *Pool) Start() {
	for i := 0; i < p.size; i++ {
		p.workers = append(p.workers, &worker{
			id:                     i,
			isWork:                 false,
			job:                    make(chan func()),
			shutdown:               make(chan struct{}),
			cumulativeNumberOfJobs: 0,
		})
	}
	wg := &sync.WaitGroup{}
	for _, w := range p.workers {
		wg.Add(1)
		go func(w *worker, wg *sync.WaitGroup) {
			wg.Done()
			fmt.Println(w.id, "start...")
			w.run()
		}(w, wg)
	}
	wg.Wait()
	fmt.Println("=all worker start=")
	for {
		select {
		case f := <-p.jobCache:
			{
				for {
					sendOk := false
					for _, w := range p.workers {
						if !w.isWork {
							w.isWork = true
							w.job <- f
							sendOk = true
							break
						}
					}
					if sendOk {
						break
					}
				}
			}

		case <-p.close:
			{
				for _, w := range p.workers {
					w.shutdown <- struct{}{}
				}
			}
		case <-time.Tick(p.timeout):
			{
				go func(p *Pool) {
					p.close <- struct{}{}
				}(p)
			}
		}
	}
}

// 往线程池添加任务
func (p *Pool) AddJob(job func()) {
	p.jobCache <- job
}

type worker struct {
	id                     int           // 线程的id
	isWork                 bool          // 是否正在工作
	job                    chan func()   // 工作时执行的函数
	shutdown               chan struct{} // 是否关闭当前协程
	cumulativeNumberOfJobs int           // 当前协程累计工作的次数
}

// 启动协程
func (w *worker) run() {
	for {
		select {
		case job := <-w.job:
			{
				fmt.Println("myId is: ", w.id)
				job()
				w.cumulativeNumberOfJobs += 1
				w.isWork = false
			}
		case <-w.shutdown:
			{
				fmt.Println(w.id, "累计完成工作次数", w.cumulativeNumberOfJobs, "Shutdown...")
				return
			}
		case <-time.Tick(1 * time.Second):
			{
				fmt.Println(w.id, "空闲")
			}
		}

	}
}
