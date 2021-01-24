// Copyright (c) 2021 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package refreshable

type Supplier interface {
	// Current returns the most recent value of this Supplier.
	Current() interface{}
}

type Refreshable interface {
	// Supplier is embedded to provide access to the most recent value of this Refreshable.
	Supplier

	// Subscribe subscribes to changes of this Refreshable. The provided function is called with the value of Current()
	// whenever the value changes.
	Subscribe(consumer func(interface{})) (unsubscribe func())

	// Map returns a new Refreshable based on the current one that handles updates based on the current Refreshable.
	Map(func(interface{}) interface{}) Refreshable
}

// A SettableRefreshable is a refreshable that can have its value updated by calling "update".
type SettableRefreshable interface {
	Refreshable

	Update(val interface{}) error
}
