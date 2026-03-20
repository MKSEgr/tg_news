package scheduler

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewValidation(t *testing.T) {
	if _, err := New(0, func(context.Context) error { return nil }); err == nil {
		t.Fatalf("expected interval validation error")
	}
	if _, err := New(time.Second, nil); err == nil {
		t.Fatalf("expected nil job validation error")
	}
	if _, err := NewWithRetry(time.Second, func(context.Context) error { return nil }, RetryPolicy{MaxAttempts: -1}); err == nil {
		t.Fatalf("expected max attempts validation error")
	}
	if _, err := NewWithRetry(time.Second, func(context.Context) error { return nil }, RetryPolicy{Backoff: -1}); err == nil {
		t.Fatalf("expected backoff validation error")
	}
}

func TestRunExecutesJobUntilCancel(t *testing.T) {
	var calls atomic.Int64
	ctx, cancel := context.WithCancel(context.Background())

	s, err := New(5*time.Millisecond, func(context.Context) error {
		if calls.Add(1) >= 2 {
			cancel()
		}
		return nil
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if err := s.Run(ctx); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if calls.Load() < 2 {
		t.Fatalf("calls = %d, want >= 2", calls.Load())
	}
}

func TestRunReturnsJobError(t *testing.T) {
	expected := errors.New("boom")
	s, err := New(time.Millisecond, func(context.Context) error { return expected })
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	err = s.Run(context.Background())
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestRunRetriesRetryableError(t *testing.T) {
	var calls atomic.Int64
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s, err := NewWithRetry(time.Millisecond, func(context.Context) error {
		if calls.Add(1) < 3 {
			return RetryableError{Err: errors.New("temporary network failure")}
		}
		cancel()
		return nil
	}, RetryPolicy{MaxAttempts: 2, Backoff: time.Millisecond})
	if err != nil {
		t.Fatalf("NewWithRetry() error = %v", err)
	}

	if err := s.Run(ctx); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if got := calls.Load(); got != 3 {
		t.Fatalf("calls = %d, want 3", got)
	}
}

func TestRunRetriesPointerRetryableError(t *testing.T) {
	var calls atomic.Int64
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s, err := NewWithRetry(time.Millisecond, func(context.Context) error {
		if calls.Add(1) < 2 {
			return &RetryableError{Err: errors.New("temporary network failure")}
		}
		cancel()
		return nil
	}, RetryPolicy{MaxAttempts: 1, Backoff: time.Millisecond})
	if err != nil {
		t.Fatalf("NewWithRetry() error = %v", err)
	}

	if err := s.Run(ctx); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if got := calls.Load(); got != 2 {
		t.Fatalf("calls = %d, want 2", got)
	}
}

func TestRunDoesNotRetryNonRetryableError(t *testing.T) {
	var calls atomic.Int64
	expected := errors.New("permanent failure")

	s, err := NewWithRetry(time.Millisecond, func(context.Context) error {
		calls.Add(1)
		return expected
	}, RetryPolicy{MaxAttempts: 3, Backoff: time.Millisecond})
	if err != nil {
		t.Fatalf("NewWithRetry() error = %v", err)
	}

	err = s.Run(context.Background())
	if err == nil {
		t.Fatalf("expected error")
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("calls = %d, want 1", got)
	}
}

func TestRunStopsWhenRetryBudgetExhausted(t *testing.T) {
	var calls atomic.Int64

	s, err := NewWithRetry(time.Millisecond, func(context.Context) error {
		calls.Add(1)
		return RetryableError{Err: errors.New("transient but persistent")}
	}, RetryPolicy{MaxAttempts: 2})
	if err != nil {
		t.Fatalf("NewWithRetry() error = %v", err)
	}

	err = s.Run(context.Background())
	if err == nil {
		t.Fatalf("expected error")
	}
	if got := calls.Load(); got != 3 {
		t.Fatalf("calls = %d, want 3", got)
	}
}

func TestRunReturnsNilWhenContextCanceledDuringBackoff(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var calls atomic.Int64

	s, err := NewWithRetry(time.Millisecond, func(context.Context) error {
		if calls.Add(1) == 1 {
			cancel()
			return RetryableError{Err: errors.New("temporary network failure")}
		}
		return nil
	}, RetryPolicy{MaxAttempts: 2, Backoff: 10 * time.Millisecond})
	if err != nil {
		t.Fatalf("NewWithRetry() error = %v", err)
	}

	if err := s.Run(ctx); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("calls = %d, want 1", got)
	}
}

func TestRunValidation(t *testing.T) {
	s, err := New(time.Millisecond, func(context.Context) error { return nil })
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if err := s.Run(nil); err == nil {
		t.Fatalf("expected nil context error")
	}

	var nilScheduler *Scheduler
	if err := nilScheduler.Run(context.Background()); err == nil {
		t.Fatalf("expected nil scheduler error")
	}

	zeroInterval := &Scheduler{job: func(context.Context) error { return nil }}
	if err := zeroInterval.Run(context.Background()); err == nil {
		t.Fatalf("expected zero interval error")
	}

	nilJob := &Scheduler{interval: time.Millisecond}
	if err := nilJob.Run(context.Background()); err == nil {
		t.Fatalf("expected nil job error")
	}

	negativeAttempts := &Scheduler{interval: time.Millisecond, job: func(context.Context) error { return nil }, retry: RetryPolicy{MaxAttempts: -1}}
	if err := negativeAttempts.Run(context.Background()); err == nil {
		t.Fatalf("expected negative attempts error")
	}

	negativeBackoff := &Scheduler{interval: time.Millisecond, job: func(context.Context) error { return nil }, retry: RetryPolicy{Backoff: -1}}
	if err := negativeBackoff.Run(context.Background()); err == nil {
		t.Fatalf("expected negative backoff error")
	}
}
