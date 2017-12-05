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

package autoscaling

import (
	. "github.com/onsi/ginkgo"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/test/e2e/framework"
)

const dummyDriverResource = "http://www.frobware.com/~aim/adapter.yaml"

var _ = SIGDescribe("[DUMMYDRIVER] Horizontal pod autoscaling (scale resource: Custom Metrics from DummyDriver)", func() {
	f := framework.NewDefaultFramework("horizontal-pod-autoscaling-dummydriver")
	It("should autoscale with Custom Metrics from DummyDriver [Feature:CustomMetricsAutoscaling]", func() {
		testHPAUsingDummyDriver(f, f.ClientSet)
	})
})

func kubectl(action, resource string) error {
	status, err := framework.RunKubectl(action, "-f", resource)
	framework.Logf(status)
	return err
}

func testHPAUsingDummyDriver(f *framework.Framework, kubeClient clientset.Interface) {
	if err := kubectl("create", dummyDriverResource); err != nil {
		framework.Failf("Failed to set up: %s", err)
	}
	defer kubectl("delete", dummyDriverResource)

	// Run application that exports the metric
	err := createDeploymentsToScale(f, kubeClient)
	if err != nil {
		framework.Failf("Failed to create stackdriver-exporter pod: %v", err)
	}
	defer cleanupDeploymentsToScale(f, kubeClient)
}
