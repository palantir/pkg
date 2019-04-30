// Copyright (c) 2019 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package typenames

import (
	"encoding/json"
	"math/big"
	"reflect"
)

func IsBigIntType(typ reflect.Type) bool {
	return typ == reflect.TypeOf(new(big.Int))
}

func IsBigFloatType(typ reflect.Type) bool {
	return typ == reflect.TypeOf(new(big.Float))
}

func IsJSONNumberType(typ reflect.Type) bool {
	return typ == reflect.TypeOf(json.Number(""))
}
