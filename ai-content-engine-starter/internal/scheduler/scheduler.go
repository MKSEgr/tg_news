package scheduler

import (
	"context"
	"fmt"
	"time"
)

// Job represents a periodic unit of work.
type Job func(ctx context.Context) error

// Scheduler runs a job on a fixed interval.
type Scheduler struct {
	interval time.Duration
	job      Job
}

// New creates a scheduler.
func New(interval time.Duration, job Job) (*Scheduler, error) {
	if interval <= 0 {
		return nil, fmt.Errorf("interval must be greater than zero")
	}
	if job == nil {
		return nil, fmt.Errorf("job is nil")
	}
	return &Scheduler{interval: interval, job: job}, nil
}

// Run starts periodic execution until context cancellation.
func (s *Scheduler) Run(ctx context.Context) error {
	if s == nil {
		return fmt.Errorf("scheduler is nil")
	}
	if ctx == nil {
		return fmt.Errorf("context is nil")
	}
	if s.interval <= 0 {
		return fmt.Errorf("interval must be greater than zero")
	}
	if s.job == nil {
		return fmt.Errorf("job is nil")
	}

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := s.job(ctx); err != nil {
				return fmt.Errorf("run job: %w", err)
			}
		}
	}
}
