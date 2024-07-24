package loadtest

import (
	"sync"
)

type WorkerPool struct {
	workerCount int
	jobs        chan func()
	wg          sync.WaitGroup
}

func NewWorkerPool(workerCount int) *WorkerPool {
	return &WorkerPool{
		workerCount: workerCount,
		jobs:        make(chan func(), workerCount),
	}
}

func (wp *WorkerPool) Start() {
	for i := 0; i < wp.workerCount; i++ {
		wp.wg.Add(1)
		go func() {
			defer wp.wg.Done()
			for job := range wp.jobs {
				job()
			}
		}()
	}
}

func (wp *WorkerPool) Submit(job func()) {
	wp.jobs <- job
}

func (wp *WorkerPool) Stop() {
	close(wp.jobs)
	wp.wg.Wait()
}
