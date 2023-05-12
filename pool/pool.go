package pool

import (
	"fmt"
	"sync"
	"time"
)

type Pool struct {
	Size    int
	Cache   chan func()
	Workers []*Worker
	Close   chan struct{}
	Timeout time.Duration
}

func (p *Pool) Start() {
	wg := &sync.WaitGroup{}
	for _, w := range p.Workers {
		wg.Add(1)
		go func(w *Worker, wg *sync.WaitGroup) {
			wg.Done()
			fmt.Println(w.Id, "start...")
			w.run()
		}(w, wg)
	}
	wg.Wait()
	fmt.Println("=all worker start=")
	for {
		select {
		case f := <-p.Cache:
			{
				for {
					sendOk := false
					for _, w := range p.Workers {
						if !w.IsWork {
							w.IsWork = true
							w.Job <- f
							sendOk = true
							break
						}
					}
					if sendOk {
						break
					}
				}

			}

		case <-p.Close:
			{
				for _, w := range p.Workers {
					w.Shutdown <- struct{}{}
				}
			}
		case <-time.Tick(p.Timeout):
			{
				go func(p *Pool) {
					p.Close <- struct{}{}
				}(p)
			}
		}
	}
}

type Worker struct {
	Id                     int
	IsWork                 bool
	Job                    chan func()
	Shutdown               chan struct{}
	Mutex                  sync.Mutex
	CumulativeNumberOfJobs int
}

func (w *Worker) run() {
	for {
		select {
		case job := <-w.Job:
			{
				fmt.Println("myId is: ", w.Id)
				job()
				w.CumulativeNumberOfJobs += 1
				w.IsWork = false
			}
		case <-w.Shutdown:
			{
				fmt.Println(w.Id, "累计完成工作次数", w.CumulativeNumberOfJobs, "Shutdown...")
				return
			}
		case <-time.Tick(1 * time.Second):
			{
				fmt.Println(w.Id, "空闲")
			}
		}

	}
}
