// Copyright 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package remoting

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid" // TODO Switch to the recommended implementation once there is one.
)

func NewErrorInstanceID() ErrorInstanceID {
	return [16]byte(uuid.New())
}

// ErrorInstanceID uniquely identifies error instance.
type ErrorInstanceID [16]byte

// String return error instance id in the UUID string form: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx".
func (eiid ErrorInstanceID) String() string {
	return uuid.UUID(eiid).String()
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

	parsed, err := uuid.Parse(eiidString)
	if err != nil {
		return fmt.Errorf("remoting: %s", err.Error())
	}

	*eiid = ErrorInstanceID(parsed)
	return nil
}
