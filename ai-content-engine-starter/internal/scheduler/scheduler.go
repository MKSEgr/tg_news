package scheduler

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Job represents a periodic unit of work.
type Job func(ctx context.Context) error

// RetryableError marks a job failure as safe to retry within the current tick.
type RetryableError struct {
	Err error
}

// Error implements error.
func (e RetryableError) Error() string {
	if e.Err == nil {
		return "retryable error"
	}
	return e.Err.Error()
}

// Unwrap returns the wrapped error.
func (e RetryableError) Unwrap() error {
	return e.Err
}

// RetryPolicy configures per-tick retries for transient external failures.
type RetryPolicy struct {
	MaxAttempts int
	Backoff     time.Duration
}

// Scheduler runs a job on a fixed interval.
type Scheduler struct {
	interval time.Duration
	job      Job
	retry    RetryPolicy
}

// New creates a scheduler.
func New(interval time.Duration, job Job) (*Scheduler, error) {
	return NewWithRetry(interval, job, RetryPolicy{})
}

// NewWithRetry creates a scheduler with an optional retry policy.
func NewWithRetry(interval time.Duration, job Job, retry RetryPolicy) (*Scheduler, error) {
	if interval <= 0 {
		return nil, fmt.Errorf("interval must be greater than zero")
	}
	if job == nil {
		return nil, fmt.Errorf("job is nil")
	}
	if retry.MaxAttempts < 0 {
		return nil, fmt.Errorf("max attempts must be greater than or equal to zero")
	}
	if retry.Backoff < 0 {
		return nil, fmt.Errorf("backoff must be greater than or equal to zero")
	}
	return &Scheduler{interval: interval, job: job, retry: retry}, nil
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
	if s.retry.MaxAttempts < 0 {
		return fmt.Errorf("max attempts must be greater than or equal to zero")
	}
	if s.retry.Backoff < 0 {
		return fmt.Errorf("backoff must be greater than or equal to zero")
	}

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := s.runTick(ctx); err != nil {
				return fmt.Errorf("run job: %w", err)
			}
		}
	}
}

func (s *Scheduler) runTick(ctx context.Context) error {
	attempts := s.retry.MaxAttempts + 1
	if attempts <= 0 {
		attempts = 1
	}
	for attempt := 1; attempt <= attempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return nil
		}
		err := s.job(ctx)
		if err == nil {
			return nil
		}
		if !isRetryable(err) || attempt == attempts {
			return err
		}
		if err := sleepContext(ctx, s.retry.Backoff); err != nil {
			return err
		}
	}
	return nil
}

func isRetryable(err error) bool {
	var retryable RetryableError
	if errors.As(err, &retryable) {
		return true
	}
	var retryablePtr *RetryableError
	return errors.As(err, &retryablePtr)
}

func sleepContext(ctx context.Context, wait time.Duration) error {
	if wait <= 0 {
		select {
		case <-ctx.Done():
			return nil
		default:
			return nil
		}
	}
	timer := time.NewTimer(wait)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return nil
	case <-timer.C:
		return nil
	}
}
