// Copyright (c) 2018 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rid

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// A ResourceIdentifier is a four-part identifier string for a resource
// whose format is specified at https://github.com/palantir/resource-identifier.
//
// Resource Identifiers offer a common encoding for wrapping existing unique
// identifiers with some additional context that can be useful when storing
// those identifiers in other applications. Additionally, the context can be
// used to disambiguate application-unique, but not globally-unique,
// identifiers when used in a common space.
type ResourceIdentifier struct {
	// Service is a string that represents the service (or application) that namespaces the rest of the identifier.
	// Must conform with regex pattern [a-z][a-z0-9\-]*.
	Service string
	// Instance is an optionally empty string that represents a specific service cluster, to allow disambiguation of artifacts from different service clusters.
	// Must conform to regex pattern ([a-z0-9][a-z0-9\-]*)?.
	Instance string
	// Type is a service-specific resource type to namespace a group of locators.
	// Must conform to regex pattern [a-z][a-z0-9\-]*.
	Type string
	// Locator is a string used to uniquely locate the specific resource.
	// Must conform to regex pattern [a-zA-Z0-9\-\._]+.
	Locator string
}

const (
	ridClass  = "ri"
	separator = "."
)

func MustNew(service, instance, resourceType, locator string) ResourceIdentifier {
	resourceIdentifier, err := New(service, instance, resourceType, locator)
	if err != nil {
		panic(err)
	}
	return resourceIdentifier
}

func New(service, instance, resourceType, locator string) (ResourceIdentifier, error) {
	resourceIdentifier := ResourceIdentifier{
		Service:  service,
		Instance: instance,
		Type:     resourceType,
		Locator:  locator,
	}
	return resourceIdentifier, resourceIdentifier.validate()
}

func (rid ResourceIdentifier) String() string {
	return ridClass + separator + rid.Service + separator + rid.Instance + separator + rid.Type + separator + rid.Locator
}

// MarshalText implements encoding.TextMarshaler (used by encoding/json and others).
func (rid ResourceIdentifier) MarshalText() (text []byte, err error) {
	return []byte(rid.String()), rid.validate()
}

// UnmarshalText implements encoding.TextUnmarshaler (used by encoding/json and others).
func (rid *ResourceIdentifier) UnmarshalText(text []byte) error {
	var err error
	parsed, err := ParseRID(string(text))
	if err != nil {
		return err
	}
	*rid = parsed
	return nil
}

// ParseRID parses a string into a 4-part resource identifier.
func ParseRID(s string) (ResourceIdentifier, error) {
	segments := strings.SplitN(s, separator, 5)
	if len(segments) != 5 || segments[0] != ridClass {
		return ResourceIdentifier{}, errors.New("invalid resource identifier")
	}
	rid := ResourceIdentifier{
		Service:  segments[1],
		Instance: segments[2],
		Type:     segments[3],
		Locator:  segments[4],
	}
	return rid, rid.validate()
}

var (
	servicePattern  = regexp.MustCompile(`^[a-z][a-z0-9\-]*$`)
	instancePattern = regexp.MustCompile(`^([a-z0-9][a-z0-9\-]*)?$`)
	typePattern     = regexp.MustCompile(`^[a-z][a-z0-9\-]*$`)
	locatorPattern  = regexp.MustCompile(`^[a-zA-Z0-9\-\._]+$`)
)

func (rid ResourceIdentifier) validate() error {
	var msgs []string
	if !servicePattern.MatchString(rid.Service) {
		msgs = append(msgs, fmt.Sprintf("rid first segment (service) does not match %s pattern", servicePattern))
	}
	if !instancePattern.MatchString(rid.Instance) {
		msgs = append(msgs, fmt.Sprintf("rid second segment (instance) does not match %s pattern", instancePattern))
	}
	if !typePattern.MatchString(rid.Type) {
		msgs = append(msgs, fmt.Sprintf("rid third segment (type) does not match %s pattern", typePattern))
	}
	if !locatorPattern.MatchString(rid.Locator) {
		msgs = append(msgs, fmt.Sprintf("rid fourth segment (locator) does not match %s pattern", locatorPattern))
	}
	if len(msgs) != 0 {
		return errors.New(strings.Join(msgs, ": "))
	}
	return nil
}
