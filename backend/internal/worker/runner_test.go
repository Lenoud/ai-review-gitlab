package worker

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/Lenoud/ai-review-gitlab/backend/internal/service"
	"github.com/stretchr/testify/require"
)

func TestRunnerProcessesImmediatelyAndStopsOnCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	processor := &fakeProcessor{
		results: []*service.ReviewWorkerResult{{Processed: true, TaskID: 1}},
		afterCall: func() {
			cancel()
		},
	}
	runner := NewRunner(processor, RunnerOptions{
		WorkerID:      "worker-1",
		IdleInterval:  time.Hour,
		ErrorInterval: time.Hour,
		Sleep:         fakeSleepNoop,
	})

	runner.Start(ctx)
	runner.Wait()

	require.Equal(t, []string{"worker-1"}, processor.workerIDs)
}

func TestRunnerSleepsAfterIdleResult(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var sleeper fakeSleeper
	processor := &fakeProcessor{
		results: []*service.ReviewWorkerResult{{Processed: false}},
		afterCall: func() {
			cancel()
		},
	}
	runner := NewRunner(processor, RunnerOptions{
		WorkerID:      "worker-1",
		IdleInterval:  5 * time.Second,
		ErrorInterval: 30 * time.Second,
		Sleep:         sleeper.sleep,
	})

	runner.Start(ctx)
	runner.Wait()

	require.Equal(t, []time.Duration{5 * time.Second}, sleeper.durations)
}

func TestRunnerSleepsAfterError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var sleeper fakeSleeper
	processor := &fakeProcessor{
		errs: []error{errors.New("boom")},
		afterCall: func() {
			cancel()
		},
	}
	runner := NewRunner(processor, RunnerOptions{
		WorkerID:      "worker-1",
		IdleInterval:  5 * time.Second,
		ErrorInterval: 30 * time.Second,
		Sleep:         sleeper.sleep,
	})

	runner.Start(ctx)
	runner.Wait()

	require.Equal(t, []time.Duration{30 * time.Second}, sleeper.durations)
}

type fakeProcessor struct {
	mu        sync.Mutex
	results   []*service.ReviewWorkerResult
	errs      []error
	workerIDs []string
	afterCall func()
}

func (p *fakeProcessor) ProcessNext(ctx context.Context, workerID string) (*service.ReviewWorkerResult, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.workerIDs = append(p.workerIDs, workerID)
	var result *service.ReviewWorkerResult
	if len(p.results) > 0 {
		result = p.results[0]
		p.results = p.results[1:]
	} else {
		result = &service.ReviewWorkerResult{Processed: false}
	}
	var err error
	if len(p.errs) > 0 {
		err = p.errs[0]
		p.errs = p.errs[1:]
	}
	if p.afterCall != nil {
		p.afterCall()
	}
	return result, err
}

type fakeSleeper struct {
	durations []time.Duration
}

func (s *fakeSleeper) sleep(ctx context.Context, d time.Duration) {
	s.durations = append(s.durations, d)
}

func fakeSleepNoop(ctx context.Context, d time.Duration) {}
