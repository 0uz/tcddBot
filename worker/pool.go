package worker

import (
	"context"
	"sync"
	"time"
)

type Job struct {
	ChatID           int64
	DepartureStation int
	ArrivalStation   int
	TravelDate       string
	LastNotified     time.Time // Add this field to track last notification
}

type Pool struct {
	workers  int
	jobQueue chan Job
	results  chan error
	wg       sync.WaitGroup
	handler  func(context.Context, Job) error
}

func NewPool(workers int, queueSize int, handler func(context.Context, Job) error) *Pool {
	return &Pool{
		workers:  workers,
		jobQueue: make(chan Job, queueSize),
		results:  make(chan error, queueSize),
		handler:  handler,
	}
}

func (p *Pool) Start(ctx context.Context) {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker(ctx)
	}
}

func (p *Pool) Stop() {
	close(p.jobQueue)
	p.wg.Wait()
	close(p.results)
}

func (p *Pool) AddJob(job Job) {
	p.jobQueue <- job
}

func (p *Pool) worker(ctx context.Context) {
	defer p.wg.Done()

	for job := range p.jobQueue {
		select {
		case <-ctx.Done():
			return
		default:
			if err := p.handler(ctx, job); err != nil {
				p.results <- err
			}
		}
	}
}
