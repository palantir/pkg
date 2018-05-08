// Copyright 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package remoting

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
)

type ErrorInstanceID [16]byte // TODO Can I use existing UUID implementation?

func (eiid ErrorInstanceID) String() string {
	return fmt.Sprintf("%x-%x-%x-%x-%x",
		eiid[0:4], eiid[4:6], eiid[6:8], eiid[8:10], eiid[10:16])
}

func (eiid ErrorInstanceID) MarshalJSON() ([]byte, error) {
	return json.Marshal(eiid.String())
}

func (eiid *ErrorInstanceID) UnmarshalJSON(data []byte) error {
	var eiidString string

	err := json.Unmarshal(data, &eiidString)
	if err != nil {
		return err
	}

	t := []byte(eiidString)
	if t[8] != '-' || t[13] != '-' || t[18] != '-' || t[23] != '-' {
		return fmt.Errorf("remoting: incorrect error instance id format %s", t)
	}

	src := []byte(eiidString)[:]
	dst := eiid[:]
	for i, byteGroup := range []int{8, 4, 4, 4, 12} {
		if i > 0 {
			src = src[1:] // skip dash
		}
		_, err = hex.Decode(dst[:byteGroup/2], src[:byteGroup])
		if err != nil {
			return fmt.Errorf("remoting: could not decode byte group %d: %s", i, src[:byteGroup])
		}
		src = src[byteGroup:]
		dst = dst[byteGroup/2:]
	}
	return nil
}
