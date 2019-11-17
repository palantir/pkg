// Copyright (c) 2018 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rid_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/palantir/pkg/rid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResourceIdentifier(t *testing.T) {
	for _, test := range []struct {
		Name        string
		Input       rid.ResourceIdentifier
		Expected    string
		ExpectedErr string
	}{
		{
			Name: "basic RID",
			Input: rid.ResourceIdentifier{
				Service:  "my-service",
				Instance: "my-instance",
				Type:     "my-type",
				Locator:  "my.locator.with.dots",
			},
			Expected: "ri.my-service.my-instance.my-type.my.locator.with.dots",
		},
		{
			Name: "empty instance",
			Input: rid.ResourceIdentifier{
				Service:  "my-service",
				Instance: "",
				Type:     "my-type",
				Locator:  "my.locator.with.dots",
			},
			Expected: "ri.my-service..my-type.my.locator.with.dots",
		},
		{
			Name: "invalid casing",
			Input: rid.ResourceIdentifier{
				Service:  "myService",
				Instance: "myInstance",
				Type:     "myType",
				Locator:  "my.locator.with.dots",
			},
			ExpectedErr: `rid first segment (service) does not match ^[a-z][a-z0-9\-]*$ pattern: rid second segment (instance) does not match ^([a-z0-9][a-z0-9\-]*)?$ pattern: rid third segment (type) does not match ^[a-z][a-z0-9\-]*$ pattern`,
		},
		{
			Name: "service can't be empty",
			Input: rid.ResourceIdentifier{
				Service:  "",
				Instance: "b",
				Type:     "c",
				Locator:  "d",
			},
			ExpectedErr: `rid first segment (service) does not match ^[a-z][a-z0-9\-]*$ pattern`,
		},
		{
			Name: "service can't contain '.'",
			Input: rid.ResourceIdentifier{
				Service:  "a.a",
				Instance: "b",
				Type:     "c",
				Locator:  "d",
			},
			ExpectedErr: `rid first segment (service) does not match ^[a-z][a-z0-9\-]*$ pattern`,
		},
		{
			Name: "instance can't contain '.'",
			Input: rid.ResourceIdentifier{
				Service:  "a",
				Instance: "b.b",
				Type:     "c",
				Locator:  "d",
			},
			ExpectedErr: `rid second segment (instance) does not match ^([a-z0-9][a-z0-9\-]*)?$ pattern`,
		},
		{
			Name: "type can't be empty",
			Input: rid.ResourceIdentifier{
				Service:  "a",
				Instance: "b",
				Type:     "",
				Locator:  "d",
			},
			ExpectedErr: `rid third segment (type) does not match ^[a-z][a-z0-9\-]*$ pattern`,
		},
		{
			Name: "type can't contain '.'",
			Input: rid.ResourceIdentifier{
				Service:  "a",
				Instance: "b",
				Type:     "c.c",
				Locator:  "d",
			},
			ExpectedErr: `rid third segment (type) does not match ^[a-z][a-z0-9\-]*$ pattern`,
		},
		{
			Name: "locator can't be empty",
			Input: rid.ResourceIdentifier{
				Service:  "a",
				Instance: "b",
				Type:     "c",
				Locator:  "",
			},
			ExpectedErr: `rid fourth segment (locator) does not match ^[a-zA-Z0-9\-\._]+$ pattern`,
		},
	} {
		t.Run(test.Name, func(t *testing.T) {
			type ridContainer struct {
				RID rid.ResourceIdentifier `json:"rid"`
			}

			// Test Marshal
			jsonBytes, err := json.Marshal(ridContainer{RID: test.Input})
			if test.ExpectedErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.ExpectedErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, fmt.Sprintf(`{"rid":%q}`, test.Expected), string(jsonBytes))

			// Test Unmarshal
			var unmarshaled ridContainer
			err = json.Unmarshal(jsonBytes, &unmarshaled)
			require.NoError(t, err, "failed to unmarshal json: %s", string(jsonBytes))
			assert.Equal(t, test.Expected, unmarshaled.RID.String())
			assert.Equal(t, test.Input, unmarshaled.RID)
		})
	}
}
