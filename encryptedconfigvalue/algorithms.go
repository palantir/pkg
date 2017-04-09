// Copyright 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package encryptedconfigvalue

import (
	"fmt"
)

type AlgorithmType string

const (
	AES = AlgorithmType("AES")
	RSA = AlgorithmType("RSA")
)

var algorithms = map[AlgorithmType]Algorithm{
	AES: aesAlgorithm,
	RSA: rsaAlgorithm,
}

func (a AlgorithmType) Definition() Algorithm {
	return algorithms[a]
}

func GetAlgorithm(val string) (AlgorithmType, error) {
	algType := AlgorithmType(val)
	_, ok := algorithms[algType]
	if !ok {
		return AlgorithmType(""), fmt.Errorf("unknown algorithm: %q", val)
	}
	return algType, nil
}
