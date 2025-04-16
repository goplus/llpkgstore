package sema

type Semaphore interface {
	Go(fn func() error)
	Wait() error
}
