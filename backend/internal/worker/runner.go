package worker

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/Lenoud/ai-review-gitlab/backend/internal/service"
)

type Processor interface {
	ProcessNext(ctx context.Context, workerID string) (*service.ReviewWorkerResult, error)
}

type SleepFunc func(ctx context.Context, d time.Duration)

type RunnerOptions struct {
	WorkerID      string
	IdleInterval  time.Duration
	ErrorInterval time.Duration
	Sleep         SleepFunc
}

type Runner struct {
	processor     Processor
	workerID      string
	idleInterval  time.Duration
	errorInterval time.Duration
	sleep         SleepFunc
	wg            sync.WaitGroup
}

func NewRunner(processor Processor, opts RunnerOptions) *Runner {
	workerID := strings.TrimSpace(opts.WorkerID)
	if workerID == "" {
		workerID = "review-worker-1"
	}
	if opts.IdleInterval <= 0 {
		opts.IdleInterval = 5 * time.Second
	}
	if opts.ErrorInterval <= 0 {
		opts.ErrorInterval = 30 * time.Second
	}
	if opts.Sleep == nil {
		opts.Sleep = sleepContext
	}
	return &Runner{
		processor:     processor,
		workerID:      workerID,
		idleInterval:  opts.IdleInterval,
		errorInterval: opts.ErrorInterval,
		sleep:         opts.Sleep,
	}
}

func (r *Runner) Start(ctx context.Context) {
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		r.run(ctx)
	}()
}

func (r *Runner) Wait() {
	r.wg.Wait()
}

func (r *Runner) run(ctx context.Context) {
	for {
		if ctx.Err() != nil {
			return
		}
		result, err := r.processor.ProcessNext(ctx, r.workerID)
		if err != nil {
			log.Printf("review worker process failed: %v", err)
			r.sleep(ctx, r.errorInterval)
			continue
		}
		if result == nil || !result.Processed {
			r.sleep(ctx, r.idleInterval)
		}
	}
}

func sleepContext(ctx context.Context, d time.Duration) {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
	case <-timer.C:
	}
}
