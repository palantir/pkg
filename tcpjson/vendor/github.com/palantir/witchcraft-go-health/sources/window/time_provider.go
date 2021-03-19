// Copyright (c) 2020 Palantir Technologies. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package window

import (
	"time"
)

// TimeProvider exists to supply a means of testing time window
// changes without actually taking the time to sleep
type TimeProvider interface {
	Now() time.Time
}

type ordinaryTimeProvider struct{}

func (o *ordinaryTimeProvider) Now() time.Time {
	return time.Now()
}

// NewOrdinaryTimeProvider creates a new time provider that returns time.Now().
func NewOrdinaryTimeProvider() TimeProvider {
	return &ordinaryTimeProvider{}
}

type offsetTimeProvider struct {
	offset time.Duration
}

func (o *offsetTimeProvider) Now() time.Time {
	return time.Now().Add(o.offset)
}

func (o *offsetTimeProvider) RestlessSleep(duration time.Duration) {
	o.offset = time.Duration(o.offset.Nanoseconds() + duration.Nanoseconds())
}
