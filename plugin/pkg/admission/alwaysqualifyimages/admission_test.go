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
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/admission"
	api "k8s.io/kubernetes/pkg/apis/core"
)

type admissionTest struct {
	pod        api.Pod
	attributes admission.Attributes
}

func testHandler(domain string) (*AlwaysQualifyImages, error) {
	d, err := NewDomain(domain)
	if err != nil {
		return nil, err
	}
	return NewAlwaysQualifyImages(d), nil
}

func imageName(domain, repo string) string {
	if domain == "" {
		return repo
	}
	return fmt.Sprintf("%s/%s", domain, repo)
}

func makePodSpec(domain string) api.PodSpec {
	return api.PodSpec{
		InitContainers: []api.Container{
			{Name: "init1", Image: imageName(domain, "busybox")},
			{Name: "init2", Image: imageName(domain, "busybox:latest")},
			{Name: "init3", Image: imageName(domain, "foo/busybox")},
			{Name: "init4", Image: imageName(domain, "foo/busybox:v1.2.3")},
		},
		Containers: []api.Container{
			{Name: "ctrl1", Image: imageName(domain, "busybox")},
			{Name: "ctrl2", Image: imageName(domain, "busybox:latest")},
			{Name: "ctrl3", Image: imageName(domain, "foo/busybox")},
			{Name: "ctrl4", Image: imageName(domain, "foo/busybox:v1.2.3")},
		},
	}
}

func makeUnQualifiedPodSpec() api.PodSpec {
	return makePodSpec("")
}

func makeQualifiedPodSpec(domain string) api.PodSpec {
	return makePodSpec(domain)
}

func makeTestPod(podSpec api.PodSpec, operation admission.Operation, subresource string) admissionTest {
	pod := api.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "123",
			Namespace: "test",
		},
		Spec: podSpec,
	}

	attributes := admission.NewAttributesRecord(
		&pod,
		nil,
		api.Kind("Pod").WithVersion("version"),
		pod.Namespace,
		pod.Name,
		api.Resource("pods").WithVersion("version"),
		subresource,
		operation,
		nil)

	return admissionTest{
		pod:        pod,
		attributes: attributes,
	}
}

func imageNames(containers []api.Container) []string {
	n := make([]string, len(containers))
	for i := range containers {
		n[i] = containers[i].Image
	}
	return n
}

func assertImageNamesEqual(t *testing.T, expected, actual []api.Container) {
	a, b := imageNames(expected), imageNames(actual)
	if !reflect.DeepEqual(a, b) {
		t.Errorf("expected %v, got %v", a, b)
	}
}

func assertPodSpecImagesEqual(t *testing.T, expected, actual api.PodSpec) {
	assertImageNamesEqual(t, expected.InitContainers, actual.InitContainers)
	assertImageNamesEqual(t, expected.Containers, actual.Containers)
}

func TestAdmissionWhereImagesAreUnqualified(t *testing.T) {
	for _, testDomain := range []string{
		"test.io",
		"localhost",
		"localhost:5000",
		"a.b.c.d.e.f",
		"a.b.c.d.e.f:5000",
	} {
		handler, err := testHandler(testDomain)
		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}

		test := makeTestPod(makeUnQualifiedPodSpec(), admission.Create, "")

		if err := handler.Admit(test.attributes); err != nil {
			t.Fatalf("Unexpected error returned from admission handler: %s", err)
		}

		assertPodSpecImagesEqual(t, makeQualifiedPodSpec(testDomain), test.pod.Spec)
	}
}

func TestAdmissionWhereImagesAreAlreadyQualified(t *testing.T) {
	// We construct a valid variant of testDomain to ensure we
	// don't see "unexpectedtest.io" materialise in the results.
	// All the images are qualified so they should remain
	// unchanged.
	testDomain := "test.io"
	handler, err := testHandler("unexpected" + testDomain)
	if err != nil {
		t.Fatalf("Unexpected error: %q: %s", testDomain, err)
	}

	test := makeTestPod(makeQualifiedPodSpec(testDomain), admission.Create, "")

	if err := handler.Admit(test.attributes); err != nil {
		t.Fatalf("Unexpected error returned from admission handler: %s", err)
	}

	assertPodSpecImagesEqual(t, makeQualifiedPodSpec(testDomain), test.pod.Spec)
}

func TestAdmissionOnPodSubresourceDoesNotModidyImageNames(t *testing.T) {
	handler, err := testHandler("test.io")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	test := makeTestPod(makeUnQualifiedPodSpec(), admission.Create, "subresource")

	if err := handler.Admit(test.attributes); err != nil {
		t.Fatalf("Unexpected error returned from admission handler: %s", err)
	}

	// We tried calling Admit() on a subresource. That will be
	// rejected by the handler so the image names should remain
	// unqualified.

	assertPodSpecImagesEqual(t, makeUnQualifiedPodSpec(), test.pod.Spec)
}

func TestAdmissionOnNonPodObject(t *testing.T) {
	handler, err := testHandler("test.io")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	// The attributes are all correct for admission because
	// Resource=="pods", there is no subresource, et al. However,
	// Object is not of type api.Pod and the plugin will only
	// operate these types. The call to Admit() should fail.

	attributes := admission.NewAttributesRecord(
		&api.ReplicationController{},
		nil,
		api.Kind("Pod").WithVersion("version"),
		"test",
		"123",
		api.Resource("pods").WithVersion("version"),
		"",
		admission.Create,
		nil)

	expected := "Resource was marked with kind Pod but was unable to be converted"

	if err := handler.Admit(attributes); err != nil {
		if err.Error() != expected {
			t.Errorf("Expected %q, got %q", expected, err)
		}
	} else {
		t.Error("Expected an error")
	}
}

func TestAdmissionOnNonPodsResource(t *testing.T) {
	handler, err := testHandler("test.io")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	pod := api.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "123",
			Namespace: "test",
		},
		Spec: makeUnQualifiedPodSpec(),
	}

	attributes := admission.NewAttributesRecord(
		&pod,
		nil,
		api.Kind("Pod").WithVersion("version"),
		pod.Namespace,
		pod.Name,
		api.Resource("not-pods").WithVersion("version"),
		"",
		admission.Create,
		nil)

	test := admissionTest{
		pod:        pod,
		attributes: attributes,
	}

	if err := handler.Admit(test.attributes); err != nil {
		t.Fatalf("Unexpected error returned from admission handler: %v", err)
	}

	// We tried calling Admit() on a non-"pods" resource. That
	// request will be ignored by the handler so the image names
	// should remain unqualified.

	assertPodSpecImagesEqual(t, makeUnQualifiedPodSpec(), test.pod.Spec)
}

func TestAdmissionOnNonCREATERequest(t *testing.T) {
	handler, err := testHandler("test.io")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	pod := api.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "123",
			Namespace: "test",
		},
		Spec: makeUnQualifiedPodSpec(),
	}

	attributes := admission.NewAttributesRecord(
		&pod,
		nil,
		api.Kind("Pod").WithVersion("version"),
		pod.Namespace,
		pod.Name,
		api.Resource("pods").WithVersion("version"),
		"",
		admission.Update,
		nil)

	test := admissionTest{
		pod:        pod,
		attributes: attributes,
	}

	if err := handler.Admit(test.attributes); err != nil {
		t.Fatalf("Unexpected error returned from admission handler: %v", err)
	}

	// We tried calling Admit() on a non-"CREATE" resource. That
	// request will be ignored by the handler so the image names
	// should remain unqualified.

	assertPodSpecImagesEqual(t, makeUnQualifiedPodSpec(), test.pod.Spec)
}
