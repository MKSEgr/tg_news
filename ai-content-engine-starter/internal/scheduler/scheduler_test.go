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
}
