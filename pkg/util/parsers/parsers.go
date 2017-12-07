/*
Copyright 2015 The Kubernetes Authors.

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

package parsers

import (
	"fmt"
	"strings"
	//  Import the crypto sha256 algorithm for the docker image parser to work
	_ "crypto/sha256"
	//  Import the crypto/sha512 algorithm for the docker image parser to work with 384 and 512 sha hashes
	_ "crypto/sha512"

	dockerref "github.com/docker/distribution/reference"
)

const (
	DefaultImageTag = "latest"
)

// ParseImageName parses a docker image string into three parts: repo, tag and digest.
// If both tag and digest are empty, a default image tag will be returned.
func ParseImageName(image string) (string, string, string, error) {
	named, err := dockerref.ParseNormalizedNamed(image)
	if err != nil {
		return "", "", "", fmt.Errorf("couldn't parse image name: %v", err)
	}

	repoToPull := named.Name()
	var tag, digest string

	tagged, ok := named.(dockerref.Tagged)
	if ok {
		tag = tagged.Tag()
	}

	digested, ok := named.(dockerref.Digested)
	if ok {
		digest = digested.Digest().String()
	}
	// If no tag was specified, use the default "latest".
	if len(tag) == 0 && len(digest) == 0 {
		tag = DefaultImageTag
	}
	return repoToPull, tag, digest, nil
}

// SplitImageName splits a docker image string into the domain
// component and path components. An empty string is returned if there
// is no domain component. This function will first validate that
// image is a valid reference, returning an error if it is not.
// Validation is done via without normalising the image.
//
// Examples inputs and results for the domain component:
//
//   "busybox"                    -> domain is ""
//   "foo/busybox"                -> domain is ""
//   "localhost/foo/busybox"      -> domain is "localhost"
//   "localhost:5000/foo/busybox" -> domain is "localhost:5000"
//   "gcr.io/busybox"             -> domain is "gcr.io"
//   "gcr.io/foo/busybox"         -> domain is "gcr.io"
//   "docker.io/busybox"          -> domain is "docker.io"
//   "docker.io/library/busybox"  -> domain is "docker.io"
func SplitImageName(image string) (string, string, error) {
	if _, err := dockerref.Parse(image); err != nil {
		return "", "", err
	}
	i := strings.IndexRune(image, '/')
	if i == -1 || (!strings.ContainsAny(image[:i], ".:") && image[:i] != "localhost") {
		return "", image, nil
	} else {
		return image[:i], image[i+1:], nil
	}
}
