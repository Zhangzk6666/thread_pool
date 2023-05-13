package pool

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

type Pool struct {
	size                   int
	jobCache               chan func()
	workerLock             sync.Mutex
	freeSize               int
	freeWorkers            []*worker
	close                  chan struct{}
	cumulativeNumberOfJobs *int32
	timeout                time.Duration // (线程池空闲时间)超时关闭线程池
	notify                 chan struct{} // 有worker空闲时,通知
}

func NewPool() *Pool {
	return &Pool{
		size:                   runtime.NumCPU(),
		jobCache:               make(chan func(), 10),
		workerLock:             sync.Mutex{},
		freeWorkers:            []*worker{},
		freeSize:               runtime.NumCPU(),
		close:                  make(chan struct{}),
		cumulativeNumberOfJobs: new(int32),
		timeout:                5 * time.Second,
		notify:                 make(chan struct{}),
	}
}
func NewPoolWithSetParameter(size int, timeout time.Duration) *Pool {
	return &Pool{
		size:                   size,
		jobCache:               make(chan func(), 10),
		workerLock:             sync.Mutex{},
		freeSize:               size,
		freeWorkers:            []*worker{},
		close:                  make(chan struct{}),
		cumulativeNumberOfJobs: new(int32),
		timeout:                timeout,
		notify:                 make(chan struct{}),
	}
}
func (p *Pool) Start() {
	if p.size <= 0 {
		p.size = runtime.NumCPU()
	}
	for i := 0; i < p.size; i++ {
		p.freeWorkers = append(p.freeWorkers, &worker{
			id:                     i,
			job:                    make(chan func()),
			shutdown:               make(chan struct{}),
			mutex:                  sync.Mutex{},
			pool:                   p,
			cumulativeNumberOfJobs: 0,
		})
	}
	wg := &sync.WaitGroup{}
	for _, w := range p.freeWorkers {
		wg.Add(1)
		go func(w *worker, wg *sync.WaitGroup) {
			wg.Done()
			w.run()
		}(w, wg)
	}
	wg.Wait()
	fmt.Println("=all worker start=", p.size)

	wg.Add(2)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		for {
			select {
			case f := <-p.jobCache:
				{
					if p.freeSize > 0 {
						p.workerLock.Lock()
						length := len(p.freeWorkers)
						w := p.freeWorkers[length-1]
						p.freeWorkers = append(p.freeWorkers[:length-1])
						p.freeSize = len(p.freeWorkers)
						p.workerLock.Unlock()
						w.job <- f
					} else {
						// fmt.Println("---------------------")
						ch := make(chan struct{})
						flag := false
						for !flag {
							select {
							case <-p.notify:
								{
									// fmt.Println("++++++++++++++++++")
									p.workerLock.Lock()
									length := len(p.freeWorkers)
									w := p.freeWorkers[length-1]
									p.freeWorkers = append(p.freeWorkers[:length-1])
									p.freeSize = len(p.freeWorkers)
									p.workerLock.Unlock()
									w.job <- f
									flag = true
								}
							case <-ch:
								{
									// fmt.Println("++++++++++++++++++")
									p.workerLock.Lock()
									length := len(p.freeWorkers)
									w := p.freeWorkers[length-1]
									p.freeWorkers = append(p.freeWorkers[:length-1])
									p.freeSize = len(p.freeWorkers)
									p.workerLock.Unlock()
									w.job <- f
									flag = true
								}
							case <-time.Tick(10 * time.Millisecond):
								{
									if p.freeSize > 0 {
										go func(ch chan struct{}) {
											ch <- struct{}{}
										}(ch)
									}
								}
							}
						}

					}
				}
			case <-time.Tick(p.timeout):
				{
					fmt.Println("stop.......")
					//  这里还要判断是否全部线程都停止工作了
					p.workerLock.Lock()
					if len(p.freeWorkers) == p.size {
						p.close <- struct{}{}
						p.workerLock.Unlock()
						return
					}
					p.workerLock.Unlock()

				}

			}
		}
	}(wg)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		for {
			select {
			case <-p.notify:
				{
					<-time.Tick(5 * time.Millisecond)
				}
			case <-p.close:
				{
					for _, w := range p.freeWorkers {
						w.shutdown <- struct{}{}
					}
					return
				}
			}
		}
	}(wg)
	wg.Wait()
}
func (p *Pool) AddJob(job func()) {
	p.jobCache <- job
}

type worker struct {
	id                     int
	job                    chan func()
	shutdown               chan struct{}
	mutex                  sync.Mutex
	pool                   *Pool
	cumulativeNumberOfJobs int32 // 累计处理的任务数
}

func (w *worker) run() {
	for {
		select {
		case job := <-w.job:
			{
				fmt.Println("myId: ", w.id, " working.....")
				job()
				w.cumulativeNumberOfJobs += 1
				w.pool.workerLock.Lock()
				w.pool.freeWorkers = append(w.pool.freeWorkers, w)
				w.pool.freeSize = len(w.pool.freeWorkers)
				w.pool.notify <- struct{}{}
				w.pool.workerLock.Unlock()
			}
		case <-w.shutdown:
			{
				atomic.AddInt32(w.pool.cumulativeNumberOfJobs, w.cumulativeNumberOfJobs)
				fmt.Println("worker id: ", w.id, ",累计完成工作次数:", w.cumulativeNumberOfJobs,
					"Shutdown...  | workers累计完成job数: ", *w.pool.cumulativeNumberOfJobs)
				return
			}
		case <-time.Tick(1 * time.Second):
			{
				fmt.Println(w.id, "空闲")
			}
		}

	}
}
