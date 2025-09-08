// Copyright (c) 2025 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package refreshable_test

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"time"

	"github.com/palantir/pkg/refreshable/v2"
)

func ExampleNew() {
	r := refreshable.New(42)
	fmt.Println(r.Current())
	// Output: 42
}

func ExampleUpdatable_Update() {
	r := refreshable.New(42)
	r.Update(100)
	fmt.Println(r.Current())
	// Output: 100
}

func ExampleRefreshable_Subscribe() {
	r := refreshable.New(42)
	stop := r.Subscribe(func(val int) {
		fmt.Println("Updated value:", val)
	})
	r.Update(100)
	stop()
	r.Update(200)
	// Output: Updated value: 42
	// Updated value: 100
}

func ExampleCached() {
	r := refreshable.New(42)
	cached, stop := refreshable.Cached(r)
	cached.Subscribe(func(val int) {
		fmt.Println("Update:", val)
	})
	fmt.Println(cached.Current())
	r.Update(100)
	fmt.Println(cached.Current())
	r.Update(100) // No update
	stop()
	r.Update(200)
	fmt.Println(cached.Current())
	// Output: Update: 42
	// 42
	// Update: 100
	// 100
	// 100
}

func ExampleView() {
	r := refreshable.New(42)
	view := refreshable.View(r, func(val int) string {
		return fmt.Sprintf("Value: %d", val)
	})
	fmt.Println(view.Current())
	r.Update(100)
	fmt.Println(view.Current())
	r.Update(100) // Duplicate update
	fmt.Println(view.Current())
	// Output: Value: 42
	// Value: 100
	// Value: 100
}

func ExampleMap() {
	r := refreshable.New(42)
	mapped, stop := refreshable.Map(r, func(val int) string {
		return fmt.Sprintf("Mapped: %d", val)
	})
	fmt.Println(mapped.Current())
	r.Update(100)
	fmt.Println(mapped.Current())
	stop()
	r.Update(200)
	fmt.Println(mapped.Current())
	// Output: Mapped: 42
	// Mapped: 100
	// Mapped: 100
}

func ExampleMapContext() {
	ctx, cancel := context.WithCancel(context.Background())
	r := refreshable.New(42)
	mapped := refreshable.MapContext(ctx, r, func(val int) string {
		return fmt.Sprintf("Mapped: %d", val)
	})
	fmt.Println(mapped.Current())
	r.Update(100)
	fmt.Println(mapped.Current())
	cancel()
	time.Sleep(time.Millisecond)
	runtime.Gosched()
	r.Update(200)
	fmt.Println(mapped.Current())
	// Output: Mapped: 42
	// Mapped: 100
	// Mapped: 100
}

func ExampleMapWithError() {
	r := refreshable.New(42)
	validated, stop, err := refreshable.MapWithError(r, func(val int) (string, error) {
		if val < 50 {
			return "", fmt.Errorf("invalid: %d", val)
		}
		return fmt.Sprintf("Valid: %d", val), nil
	})
	if err != nil {
		fmt.Println("Initial error:", err)
	}
	fmt.Println(validated.Validation())
	r.Update(100)
	fmt.Println(validated.Validation())
	r.Update(24)
	fmt.Println(validated.Validation())
	stop()
	r.Update(200)
	fmt.Println(validated.Validation())
	// Output: Initial error: invalid: 42
	//  invalid: 42
	// Valid: 100 <nil>
	//  invalid: 24
	//  invalid: 24
}

func ExampleValidate() {
	r := refreshable.New(42)
	validated, stop, err := refreshable.Validate(r, func(val int) error {
		if val < 50 {
			return errors.New("value too low")
		}
		return nil
	})
	if err != nil {
		fmt.Println("Initial error:", err)
	}
	fmt.Println(validated.Validation())
	r.Update(100)
	fmt.Println(validated.Validation())
	stop()
	r.Update(200)
	fmt.Println(validated.Validation())
	// Output: Initial error: value too low
	// 42 value too low
	// 100 <nil>
	// 100 <nil>
}

func ExampleMerge() {
	r1 := refreshable.New(42)
	r2 := refreshable.New(100)
	merged, stop := refreshable.Merge(r1, r2, func(v1, v2 int) string {
		return fmt.Sprintf("Sum: %d", v1+v2)
	})
	fmt.Println(merged.Current())
	r1.Update(50)
	fmt.Println(merged.Current())
	r2.Update(150)
	fmt.Println(merged.Current())
	stop()
	r1.Update(60)
	fmt.Println(merged.Current())
	// Output: Sum: 142
	// Sum: 150
	// Sum: 200
	// Sum: 200
}

func ExampleCollect() {
	r1 := refreshable.New(10)
	r2 := refreshable.New(20)
	r3 := refreshable.New(30)

	collected, stop := refreshable.Collect(r1, r2, r3)

	printCollected := func() {
		values := collected.Current()
		fmt.Printf("Collected values: %v\n", values)
	}

	printCollected() // Initial values

	r1.Update(15)
	printCollected() // After updating r1

	r2.Update(25)
	printCollected() // After updating r2

	r3.Update(35)
	printCollected() // After updating r3

	stop() // Stop collecting updates

	r1.Update(40)
	printCollected() // No change after stopping

	// Output:
	// Collected values: [10 20 30]
	// Collected values: [15 20 30]
	// Collected values: [15 25 30]
	// Collected values: [15 25 35]
	// Collected values: [15 25 35]
}
