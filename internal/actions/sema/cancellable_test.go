package sema

import (
	"context"
	"errors"
	"runtime"
	"sync/atomic"
	"testing"
)

func TestCancellableGroup(t *testing.T) {
	t.Run("no-error", func(t *testing.T) {
		var cnt atomic.Int32

		inputTest := []string{"a", "b", "c", "d", "e", "f", "g"}

		g := NewSemaphoreGroup(context.TODO(), len(inputTest))

		for range inputTest {
			g.Go(func() error {
				cnt.Add(1)
				return nil
			})
		}

		err := g.Wait()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if cnt.Load() != int32(len(inputTest)) {
			t.Errorf("unexpected counter: want %d got %d", cnt.Load(), len(inputTest))
		}
	})

	t.Run("no-error-large", func(t *testing.T) {
		var cnt atomic.Int32

		inputTest := make([]byte, 10240)

		g := NewSemaphoreGroup(context.TODO(), len(inputTest))

		for range inputTest {
			g.Go(func() error {
				cnt.Add(1)
				return nil
			})
		}

		err := g.Wait()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if cnt.Load() != int32(len(inputTest)) {
			t.Errorf("unexpected counter: want %d got %d", cnt.Load(), len(inputTest))
		}
	})
	t.Run("one-error", func(t *testing.T) {
		var cnt atomic.Int32

		inputTest := []string{"a", "b", "c", "d", "e", "f", "g"}

		g := NewSemaphoreGroup(context.TODO(), len(inputTest))

		for _, x := range inputTest {
			x := x
			g.Go(func() error {
				defer cnt.Add(1)
				if x == "a" {
					return errors.New(x)
				}
				return nil
			})
		}

		err := g.Wait()
		if err.Error() != "a" {
			t.Errorf("unexpected error: %v want: a", err.Error())
		}

		if cnt.Load() == int32(len(inputTest)) {
			t.Errorf("unexpected full counter: counter should less than %d", len(inputTest))
		}
	})

	t.Run("multi-error", func(t *testing.T) {
		var cnt atomic.Int32

		inputTest := []string{"a", "b", "c", "d", "e", "f", "g"}

		g := NewSemaphoreGroup(context.TODO(), len(inputTest))

		for _, x := range inputTest {
			x := x
			g.Go(func() error {
				defer cnt.Add(1)
				if x == "a" || x == "c" {
					return errors.New(x)
				}
				return nil
			})
		}

		err := g.Wait()
		if err.Error() != "a" && err.Error() != "c" {
			t.Errorf("unexpected error: %v want: a", err)
		}

		if cnt.Load() == int32(len(inputTest)) {
			t.Errorf("unexpected full counter: counter should less than %d", len(inputTest))
		}
	})

	t.Run("parent-cancel", func(t *testing.T) {
		var cnt atomic.Int32

		inputTest := []string{"a", "b", "c", "d", "e", "f", "g"}

		parentCtx, cancel := context.WithCancel(context.TODO())
		g := NewSemaphoreGroup(parentCtx, len(inputTest))

		for range inputTest {
			cancel()
			g.Go(func() error {
				defer cnt.Add(1)
				return nil
			})
		}

		err := g.Wait()
		if err != context.Canceled {
			t.Errorf("unexpected error: %v", err)
		}
		if cnt.Load() != 0 {
			t.Errorf("unexpected counter: want %d got %d", cnt.Load(), len(inputTest))
		}
	})

	t.Run("go-size-greater-than-group-size", func(t *testing.T) {
		inputTest := []string{"a", "b", "c", "d", "e", "f", "g"}

		g := NewSemaphoreGroup(context.TODO(), 1)

		for range inputTest {
			g.Go(func() error {
				return nil
			})
		}
		runtime.Gosched()
		runtime.Gosched()

		err := g.Wait()
		if err == nil || err != ErrInvalidSize {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("zero-size", func(t *testing.T) {
		inputTest := []string{"a", "b", "c", "d", "e", "f", "g"}

		g := NewSemaphoreGroup(context.TODO(), 0)

		for range inputTest {
			g.Go(func() error {
				return nil
			})
		}
		runtime.Gosched()
		err := g.Wait()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}
