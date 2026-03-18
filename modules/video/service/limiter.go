package service

import (
	"context"

	"golang.org/x/sync/semaphore"
)

type MergeLimiter struct {
	sem *semaphore.Weighted
}

func NewMergeLimiter(maxConcurrent int64) *MergeLimiter {
	return &MergeLimiter{
		sem: semaphore.NewWeighted(maxConcurrent),
	}
}

func (l *MergeLimiter) Acquire(ctx context.Context) error {
	return l.sem.Acquire(ctx, 1)
}

func (l *MergeLimiter) Release() {
	l.sem.Release(1)
}
