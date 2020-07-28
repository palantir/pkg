// Copyright (c) 2018 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package retry

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestDo_WithMaxAttempts(t *testing.T) {
	options := []Option{
		WithInitialBackoff(time.Microsecond * 10),
		WithMaxBackoff(time.Second),
		WithMaxAttempts(3),
	}
	const expectedAttempts = 3 // First attempt and two retries.
	expectedErr := fmt.Errorf("placeholder")

	attempts := 0
	actualErr := Do(
		context.Background(),
		func() error {
			attempts++
			return expectedErr
		},
		options...,
	)
	if actualErr != expectedErr {
		t.Fatalf("expected err %v, got %v", expectedErr, actualErr)
	}
	if attempts != expectedAttempts {
		t.Errorf("expected %d attempts, got %d attempts", expectedAttempts, attempts)
	}
}

func TestDo_ReturnsImmediatelyForCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	go cancel()
	<-ctx.Done()

	actualErr := Do(ctx,
		func() error {
			t.Fatalf("action should never be invoked")
			return nil
		},
		WithInitialBackoff(time.Microsecond*10),
		WithMaxBackoff(time.Second),
	)

	if actualErr != context.Canceled {
		t.Fatalf("expected err %v, got %v", context.Canceled, actualErr)
	}
}

func TestRetrier_Next_WithMaxBackoff(t *testing.T) {
	const maxBackoff = time.Microsecond * 100

	options := []Option{
		WithInitialBackoff(time.Microsecond * 10),
		WithMaxBackoff(maxBackoff),
		WithMultiplier(2),
		WithMaxAttempts(11),
		WithRandomizationFactor(0),
	}

	r := Start(context.Background(), options...).(*retrier)
	// With 10 attempts we would easily exceed max backoff if it was not set.
	for i := 0; i < 10; i++ {
		d := r.retryIn()
		if d > maxBackoff {
			t.Fatalf("expected backoff less than max-backoff: %s vs %s", d, maxBackoff)
		}
		r.currentAttempt++
	}
}

func TestRetrier_Next_WithInitialBackoffLargerThanMax(t *testing.T) {
	const initialBackoff = time.Second * 5
	const maxBackoff = time.Second * 2

	options := []Option{
		WithInitialBackoff(initialBackoff),
		WithMaxBackoff(maxBackoff),
		WithMultiplier(2),
		WithMaxAttempts(11),
		WithRandomizationFactor(0),
	}

	r := Start(context.Background(), options...).(*retrier)
	d := r.retryIn()
	if d != initialBackoff {
		t.Fatalf("expected initial backoff to be used: %s vs %s", d, initialBackoff)
	}
}

func TestRetrier_Next_WithInitialBackoffLargerThanMaxAndMaxNotSet(t *testing.T) {
	const initialBackoff = time.Second * 5
	const maxBackoff = 0
	const multiplier = 2

	options := []Option{
		WithInitialBackoff(initialBackoff),
		WithMaxBackoff(maxBackoff),
		WithMultiplier(multiplier),
		WithMaxAttempts(11),
		WithRandomizationFactor(0),
	}

	r := Start(context.Background(), options...).(*retrier)
	r.currentAttempt = 1
	d := r.retryIn()
	if d != initialBackoff*multiplier {
		t.Fatalf("expected second attempt to increase backoff: %s vs %s", d, initialBackoff*multiplier)
	}
}

func TestRetrier_Next_WithMaxAttempts(t *testing.T) {
	const maxAttempts = 2

	options := []Option{
		WithInitialBackoff(time.Microsecond * 10),
		WithMaxBackoff(time.Second),
		WithMultiplier(2),
		WithMaxAttempts(2),
	}

	attempts := 0
	for r := Start(context.Background(), options...); r.Next(); attempts++ {
	}

	if attempts != maxAttempts {
		t.Errorf("expected %d attempts, got %d attempts", maxAttempts, attempts)
	}
}

func TestRetrier_Next_CurrentAttempt(t *testing.T) {
	r := Start(
		context.Background(),
		WithInitialBackoff(time.Microsecond*10),
		WithMaxBackoff(time.Second),
		WithMaxAttempts(2),
	)

	// First attempt.
	if next := r.Next(); !next {
		t.Errorf("cannot do first attempt")
	}
	if currentAttempt := r.CurrentAttempt(); currentAttempt != 0 {
		t.Errorf("expected current attempt to be 0, but was %d", currentAttempt)
	}

	// Second attempt.
	if next := r.Next(); !next {
		t.Errorf("cannot do final attempt")
	}
	if currentAttempt := r.CurrentAttempt(); currentAttempt != 1 {
		t.Errorf("expected current attempt to be 1, but was %d", currentAttempt)
	}

	// Another attempt call after retry is done.
	if next := r.Next(); next {
		t.Errorf("can do too many attempts")
	}
	if currentAttempt := r.CurrentAttempt(); currentAttempt != 1 {
		t.Errorf("expected current attempt to be equal to max attempts, but was %d", currentAttempt)
	}

	// First attempt after reset.
	r.Reset()
	if next := r.Next(); !next {
		t.Errorf("cannot do first attempt")
	}
	if currentAttempt := r.CurrentAttempt(); currentAttempt != 0 {
		t.Errorf("expected current attempt to be 0, but was %d", currentAttempt)
	}
}

func TestRetrier_Next_StopsOnContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	options := []Option{
		WithInitialBackoff(time.Second),
		WithMaxBackoff(time.Second),
		WithMultiplier(2),
	}

	var attempts int
	// Create a retry loop which will never stop without stopper.
	for r := Start(ctx, options...); r.Next(); attempts++ {
		go cancel()
		// Don't race the stopper, just wait for it to do its thing.
		<-ctx.Done()
	}

	if attempts != 1 {
		t.Errorf("expected 1 attempt, got %d", attempts)
	}
}
