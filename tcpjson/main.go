// +build module

// This file exists only to smooth the transition for modules. Having this file makes it such that other modules that
// consume this module will not have import path conflicts caused by github.com/palantir/pkg.
package main

import (
	_ "github.com/palantir/pkg"
)
