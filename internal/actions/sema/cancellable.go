package sema

import (
	"context"
	"errors"
	"sync/atomic"
)

var ErrInvalidSize = errors.New("sema: invalid semaphore size")

type semaCancellableGroup struct {
	sema atomic.Int32

	ctx    context.Context
	cancel context.CancelCauseFunc
}

func NewSemaphoreGroup(ctx context.Context, size int) Semaphore {
	group := &semaCancellableGroup{}
	group.sema.Add(-int32(size))

	group.ctx, group.cancel = context.WithCancelCause(ctx)

	if size == 0 {
		group.cancel(nil)
	}

	return group
}

func (s *semaCancellableGroup) Go(fn func() error) {
	select {
	case <-s.ctx.Done():
		return
	default:
	}

	go func() {
		err := fn()

		sema := s.sema.Add(1)

		if err != nil {
			s.cancel(err)
			return
		}
		if sema >= 0 {
			s.cancel(nil)
		}
	}()
}

func (s *semaCancellableGroup) Wait() error {
	<-s.ctx.Done()

	err := context.Cause(s.ctx)

	sema := s.sema.Load()
	if sema == 0 && err == context.Canceled {
		return nil
	} else if sema > 0 {
		return ErrInvalidSize
	}
	return err
}
