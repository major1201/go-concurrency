package concurrency

import "time"

// Payload the job exactly be run
type Payload struct {
	f       func()
	weight  int
	timeout time.Duration
	cancel  func()
}

// SetWeight set the payload weight
func (p *Payload) SetWeight(weight int) {
	p.weight = weight
}

// SetTimeout set the payload timeout and che cancel function,
//   caution: set a job with timeout DOES NOT make sure the job to cancel, the cancel function just be called after timeout
func (p *Payload) SetTimeout(timeout time.Duration, cancel func()) {
	p.timeout = timeout
	p.cancel = cancel
}

// NewPayload returns a payload with the job function
func NewPayload(f func()) *Payload {
	return &Payload{
		f:      f,
		weight: 1,
	}
}
