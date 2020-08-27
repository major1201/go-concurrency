package concurrency

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
	"sync"
	"time"
)

var (
	// ErrBacklogIsNotEnough indicates the backlog is not enough
	ErrBacklogIsNotEnough = errors.New("concurrency backlog is not enough")
	// ErrAlreadyQuit indicates the job group is quit
	ErrAlreadyQuit = errors.New("jobGroup is ready quit")
)

// JobGroup job group
type JobGroup struct {
	workerSize  int
	backlogSize int

	ctx        context.Context
	eg         *errgroup.Group
	mu         sync.Mutex
	sem        *semaphore.Weighted
	payloads   chan *Payload
	currWeight int // since len(channel) can't be trusted, we should use an additional integer to get the real payload weight running or in backlog

	abort chan bool
	quit  bool
}

// NewJobGroup returns a JobGroup
func NewJobGroup(workerSize, backlogSize int) *JobGroup {
	jg := &JobGroup{
		workerSize:  workerSize,
		backlogSize: backlogSize,
	}

	jg.ctx = context.Background()
	jg.eg = &errgroup.Group{}
	jg.sem = semaphore.NewWeighted(int64(workerSize))
	jg.payloads = make(chan *Payload, backlogSize+workerSize)
	jg.abort = make(chan bool, 1)
	return jg
}

// Start to consume the jobs
func (jg *JobGroup) Start() {
	for {
		select {
		case <-jg.abort:
			return
		case payload := <-jg.payloads:
			if err := jg.sem.Acquire(jg.ctx, int64(payload.weight)); err != nil {
				fmt.Printf("%+v\n", errors.Wrap(err, "semaphore acquire error"))
				break
			}
			jg.eg.Go(func() error {
				jg.runPayload(payload)
				return nil
			})
		}
	}
}

func (jg *JobGroup) runPayload(payload *Payload) {
	f := func() {
		defer func() {
			jg.mu.Lock()
			defer jg.mu.Unlock()

			jg.currWeight -= payload.weight
			jg.sem.Release(int64(payload.weight))

			if r := recover(); r != nil {
				if re, ok := r.(error); ok {
					fmt.Printf("%+v\n", errors.WithMessage(re, "[Recover from Panic]"))
				} else {
					fmt.Printf("[Recover from Panic]%v\n", r)
				}
				jg.sem.Release(int64(payload.weight))
			}
		}()

		payload.f()
	}

	if payload.timeout == 0 {
		f()
	} else {
		// run with timeout
		ch := make(chan bool)
		go func() {
			f()

			defer func() {
				recover() // recover if the channel is closed
			}()
			ch <- true
		}()

		select {
		case <-ch:
			close(ch)
		case <-time.After(payload.timeout):
			close(ch)
			if payload.cancel != nil {
				payload.cancel()
			}
		}
	}
}

// GoWithPayload commits a raw payload
func (jg *JobGroup) GoWithPayload(p *Payload) error {
	if p.weight < 1 || p.weight > jg.workerSize {
		return errors.Errorf("weight(%d) < 1 or > worker size(%d)", p.weight, jg.workerSize)
	}

	jg.mu.Lock()
	defer jg.mu.Unlock()

	if jg.quit {
		return ErrAlreadyQuit
	}

	if jg.currWeight+p.weight > jg.workerSize+jg.backlogSize {
		return ErrBacklogIsNotEnough
	}

	jg.currWeight += p.weight
	jg.payloads <- p

	return nil
}

// GoWithWeight commits a job to the backlog with weight
func (jg *JobGroup) GoWithWeight(weight int, f func()) error {
	p := NewPayload(f)
	p.SetWeight(weight)
	return jg.GoWithPayload(p)
}

// Go commits a job to the backlog
func (jg *JobGroup) Go(f func()) error {
	return jg.GoWithPayload(NewPayload(f))
}

// Abort the job group and returns at once, however the workers would still do its jobs
func (jg *JobGroup) Abort() {
	jg.mu.Lock()
	defer jg.mu.Unlock()
	jg.abort <- true
	jg.quit = true
}

// AbortWait the job group and wait until the workers is done
func (jg *JobGroup) AbortWait() {
	jg.Abort()

	_ = jg.eg.Wait()
	return
}

// StopWait the job group would no longer accept new jobs and would do and wait until all the jobs in backlog is done
func (jg *JobGroup) StopWait() {
	jg.mu.Lock()
	jg.quit = true
	jg.mu.Unlock()

	var quitAfterWait bool
	for {
		if jg.currWeight == 0 {
			quitAfterWait = true
		}
		_ = jg.eg.Wait()

		if quitAfterWait {
			return
		}
	}
}
