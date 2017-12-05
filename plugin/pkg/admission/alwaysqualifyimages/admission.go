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
	"fmt"
	"io"

	"github.com/golang/glog"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apiserver/pkg/admission"
	api "k8s.io/kubernetes/pkg/apis/core"
	"k8s.io/kubernetes/pkg/util/parsers"
)

// Register registers a plugin.
func Register(plugins *admission.Plugins) {
	plugins.Register("AlwaysQualifyImages", func(config io.Reader) (admission.Interface, error) {
		// XXX how and where does:
		//   a) versioned configuration work?!
		//   b) when is the default available?
		domain, err := NewDomain("localhost:5000")
		if err != nil {
			return nil, err
		}
		return NewAlwaysQualifyImages(domain), nil
	})
}

// AlwaysQualifyImages is an implementation of admission.Interface. It
// looks at all new pods and overrides any container's image name that
// is unqualified with Domain.
type AlwaysQualifyImages struct {
	*admission.Handler
	Domain
}

var _ admission.MutationInterface = &AlwaysQualifyImages{}

func hasDomain(image string) bool {
	domain, _, err := parsers.SplitImageName(image)
	return err == nil && domain != ""
}

// qualifyContainerImages modifies containers if the image name is
// unqualified (i.e., has no domain) to include domain. It fails fast
// if adding domain results in an invalid image.
func qualifyContainerImages(domain Domain, containers []api.Container) (string, error) {
	for i := range containers {
		if hasDomain(containers[i].Image) {
			glog.V(2).Infof("not qualifying image %q as it has a domain", containers[i].Image)
			continue
		}
		newName := string(domain) + "/" + containers[i].Image
		if _, _, _, err := parsers.ParseImageName(newName); err != nil {
			return newName, err
		}
		glog.V(2).Infof("qualifying image %q as %q", containers[i].Image, newName)
		containers[i].Image = newName
	}
	return "", nil
}

// Admit makes an admission decision based on the request attributes
func (a *AlwaysQualifyImages) Admit(attributes admission.Attributes) (err error) {
	// Ignore all calls to subresources or resources other than pods.
	// Ignore all operations other than CREATE.
	if shouldIgnore(attributes) {
		return nil
	}

	pod, ok := attributes.GetObject().(*api.Pod)
	if !ok {
		return apierrors.NewBadRequest("Resource was marked with kind Pod but was unable to be converted")
	}

	if image, err := qualifyContainerImages(a.Domain, pod.Spec.InitContainers); err != nil {
		return apierrors.NewBadRequest(fmt.Sprintf("invalid image name %q: %s", image, err.Error()))
	}

	if image, err := qualifyContainerImages(a.Domain, pod.Spec.Containers); err != nil {
		return apierrors.NewBadRequest(fmt.Sprintf("invalid image name %q: %s", image, err.Error()))
	}

	return nil
}

func isSubresourceRequest(attributes admission.Attributes) bool {
	return len(attributes.GetSubresource()) > 0
}

func isPodsRequest(attributes admission.Attributes) bool {
	return attributes.GetResource().GroupResource() == api.Resource("pods")
}

func isCreateRequest(attributes admission.Attributes) bool {
	return attributes.GetOperation() == admission.Create
}

func shouldIgnore(attributes admission.Attributes) bool {
	switch {
	case isSubresourceRequest(attributes):
		return true
	case !isPodsRequest(attributes):
		return true
	case !isCreateRequest(attributes):
		return true
	default:
		return false
	}
}

// NewAlwaysQualifyImages creates a new admission control handler that
// handled Create and Update operations and will add domain to
// unqualified Pod container image names.
func NewAlwaysQualifyImages(domain Domain) *AlwaysQualifyImages {
	return &AlwaysQualifyImages{
		Handler: admission.NewHandler(admission.Create),
		Domain:  domain,
	}
}
