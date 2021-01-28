// Copyright (c) 2021 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package refreshable

type Refreshable interface {
	// Current returns the most recent value of this Refreshable.
	Current() interface{}

	// Subscribe subscribes to changes of this Refreshable. The provided function is called with the value of Current()
	// whenever the value changes.
	Subscribe(consumer func(interface{})) (unsubscribe func())

	// Map returns a new Refreshable based on the current one that handles updates based on the current Refreshable.
	Map(func(interface{}) interface{}) Refreshable
}
