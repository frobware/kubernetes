/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package alwaysqualifyimages

import (
	"errors"

	"k8s.io/kubernetes/pkg/util/parsers"
)

type Domain string

const sanityRepo = "foo/bar:latest"

// NewDomain constructs and validates that domain can be used as the
// domain URL for a docker image reference. Returns an error if input
// would be invalid.
func NewDomain(input string) (Domain, error) {
	domain, remainder, err := parsers.SplitImageName(input + "/" + sanityRepo)
	if err != nil {
		return "", err
	}
	if domain != input && remainder != sanityRepo {
		return "", errors.New("invalid domain")
	}
	return Domain(input), nil
}
