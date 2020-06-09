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
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewBackoff(t *testing.T) {
	testCases := map[string]struct {
		wantMin, wantMax time.Duration
		gotMin, gotMax   time.Duration
	}{
		"defaults": {
			gotMin: minBackoff,
			gotMax: maxBackoff,
		},
		"all correct": {
			wantMin: time.Second,
			gotMin:  time.Second,
			wantMax: time.Minute,
			gotMax:  time.Minute,
		},
		"min > default max": {
			wantMin: time.Hour,
			gotMin:  minBackoff,
			gotMax:  maxBackoff,
		},
		"min > max": {
			wantMin: time.Hour,
			gotMin:  minBackoff,
			wantMax: time.Minute,
			gotMax:  maxBackoff,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			bo := NewBackoff()
			switch {
			case tc.wantMin != 0 && tc.wantMax != 0:
				bo = NewBackoff(tc.wantMin, tc.wantMax)
			case tc.wantMin != 0:
				bo = NewBackoff(tc.wantMin)
			}
			assert.Equal(t, bo.min, tc.gotMin, "backoff min duration has unexpected value")
			assert.Equal(t, bo.max, tc.gotMax, "backoff max duration has unexpected value")
		})
	}
}

func TestDuration(t *testing.T) {
	testCases := map[string]struct {
		step              int
		min, max, wantDur time.Duration
	}{
		"first step": {
			step:    0,
			min:     time.Second,
			max:     time.Minute,
			wantDur: time.Second,
		},
		"third step": {
			step:    2,
			min:     time.Second,
			wantDur: 3 * time.Second,
		},
		"tenth step": {
			step:    9,
			min:     time.Second,
			max:     time.Minute,
			wantDur: time.Minute,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			bo := NewBackoff()
			switch {
			case tc.min != 0 && tc.max != 0:
				bo = NewBackoff(tc.min, tc.max)
			case tc.min != 0:
				bo = NewBackoff(tc.min)
			}

			bo.step = tc.step
			assert.Equal(t, tc.wantDur, bo.Duration())
		})
	}
}

func TestRun(t *testing.T) {
	var counter int
	start := time.Now()

	testCases := map[string]struct {
		fn         func(context.Context) (bool, error)
		waitReturn bool
		err        error
	}{
		"force stop with no errs": {
			waitReturn: false,
		},
		"fn returns err": {
			fn: func(ctx context.Context) (bool, error) {
				return false, assert.AnError
			},
			waitReturn: true,
			err:        assert.AnError,
		},
		"fn resets backoff duration": {
			fn: func(ctx context.Context) (bool, error) {
				counter++
				if counter > 3 {
					if time.Since(start).Round(time.Second) != 3*time.Second {
						return true, fmt.Errorf("did we reset the backoff?")
					}
					// trigger Run() termination with expected error
					return true, assert.AnError
				}
				// reset backoff duration to 1 sec for each iteration
				return true, nil
			},
			waitReturn: true,
			err:        assert.AnError,
		},
	}

	errCh := make(chan error)
	defer close(errCh)

	for name, tc := range testCases {
		stopCh := make(chan struct{})
		t.Run(name, func(t *testing.T) {
			bo := NewBackoff()
			go func() {
				err := bo.Run(stopCh, tc.fn)
				errCh <- err
			}()
			if tc.waitReturn {
				defer close(stopCh)
			} else {
				close(stopCh)
			}

			err := <-errCh
			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
