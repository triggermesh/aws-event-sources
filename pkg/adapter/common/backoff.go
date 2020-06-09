/*
Copyright (c) 2020 TriggerMesh Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package common

import (
	"context"
	"math"
	"time"
)

// default values for backoff
const (
	expFactor = 2

	minBackoff = 1 * time.Second
	maxBackoff = 32 * time.Second
)

// Backoff provides a simple exponential backoff mechanism
type Backoff struct {
	step     int
	factor   float64
	min, max time.Duration
}

// RunFunc is a user function that polls data from a source and sends it as
// a cloudevent to a sink.
// RunFunc must return (bool, error) values where bool is true if poll backoff duration
// must be reset and error is the result of function execution.
type RunFunc func(context.Context) (bool, error)

// NewBackoff accepts optional values for minimum and maximum wait period
// and return new instance of Backoff structure
func NewBackoff(args ...time.Duration) *Backoff {
	backoff := &Backoff{
		step:   0,
		factor: expFactor,
		min:    minBackoff,
		max:    maxBackoff,
	}

	switch len(args) {
	case 1:
		if args[0] <= backoff.max {
			backoff.min = args[0]
		}
	case 2:
		if args[0] <= args[1] {
			backoff.min = args[0]
			backoff.max = args[1]
		}
	}

	return backoff
}

// Duration can be used to get exponential backoff duration calculated for each new step
func (b *Backoff) Duration() time.Duration {
	dur := time.Duration(float64(b.min)*math.Pow(b.factor, float64(b.step)) - float64(1*time.Second))

	switch {
	case dur < b.min:
		b.step++
		return b.min
	case dur > b.max:
		return b.max
	default:
		b.step++
		return dur
	}
}

// Run is a blocking function that executes RunFunc until stopCh receives the value
// or function returns an error
func (b *Backoff) Run(stopCh <-chan struct{}, fn RunFunc) error {
	timer := time.NewTimer(0)
	defer timer.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for {
		select {
		case <-stopCh:
			return nil
		case <-timer.C:
			reset, err := fn(ctx)
			if err != nil {
				return err
			}
			if reset {
				b.Reset()
			}
			timer.Reset(b.Duration())
		}
	}
}

// Reset sets step counter to zero.
func (b *Backoff) Reset() {
	b.step = 0
}
