/*
Copyright 2016 The Kubernetes Authors.

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

package rest

import (
	extensionsapiv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/rest"
	genericapiserver "k8s.io/apiserver/pkg/server"
	serverstorage "k8s.io/apiserver/pkg/server/storage"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"
	expcontrollerstore "k8s.io/kubernetes/pkg/registry/extensions/controller/storage"
	daemonstore "k8s.io/kubernetes/pkg/registry/extensions/daemonset/storage"
	deploymentstore "k8s.io/kubernetes/pkg/registry/extensions/deployment/storage"
	ingressstore "k8s.io/kubernetes/pkg/registry/extensions/ingress/storage"
	networkpolicystore "k8s.io/kubernetes/pkg/registry/extensions/networkpolicy/storage"
	pspstore "k8s.io/kubernetes/pkg/registry/extensions/podsecuritypolicy/storage"
	replicasetstore "k8s.io/kubernetes/pkg/registry/extensions/replicaset/storage"
)

type RESTStorageProvider struct {
}

func (p RESTStorageProvider) NewRESTStorage(apiResourceConfigSource serverstorage.APIResourceConfigSource, restOptionsGetter generic.RESTOptionsGetter, stopCh <-chan struct{}) (genericapiserver.APIGroupInfo, bool) {
	apiGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(extensions.GroupName, api.Registry, api.Scheme, api.ParameterCodec, api.Codecs)
	// If you add a version here, be sure to add an entry in `k8s.io/kubernetes/cmd/kube-apiserver/app/aggregator.go with specific priorities.
	// TODO refactor the plumbing to provide the information in the APIGroupInfo

	if apiResourceConfigSource.AnyResourcesForVersionEnabled(extensionsapiv1beta1.SchemeGroupVersion) {
		apiGroupInfo.VersionedResourcesStorageMap[extensionsapiv1beta1.SchemeGroupVersion.Version] = p.v1beta1Storage(apiResourceConfigSource, restOptionsGetter, stopCh)
		apiGroupInfo.GroupMeta.GroupVersion = extensionsapiv1beta1.SchemeGroupVersion
	}

	return apiGroupInfo, true
}

func (p RESTStorageProvider) v1beta1Storage(apiResourceConfigSource serverstorage.APIResourceConfigSource, restOptionsGetter generic.RESTOptionsGetter, stopCh <-chan struct{}) map[string]rest.Storage {
	version := extensionsapiv1beta1.SchemeGroupVersion

	storage := map[string]rest.Storage{}

	// This is a dummy replication controller for scale subresource purposes.
	// TODO: figure out how to enable this only if needed as a part of scale subresource GA.
	controllerStorage := expcontrollerstore.NewStorage(restOptionsGetter, stopCh)
	storage["replicationcontrollers"] = controllerStorage.ReplicationController
	storage["replicationcontrollers/scale"] = controllerStorage.Scale

	if apiResourceConfigSource.ResourceEnabled(version.WithResource("daemonsets")) {
		daemonSetStorage, daemonSetStatusStorage := daemonstore.NewREST(restOptionsGetter, stopCh)
		storage["daemonsets"] = daemonSetStorage
		storage["daemonsets/status"] = daemonSetStatusStorage
	}
	if apiResourceConfigSource.ResourceEnabled(version.WithResource("deployments")) {
		deploymentStorage := deploymentstore.NewStorage(restOptionsGetter, stopCh)
		storage["deployments"] = deploymentStorage.Deployment
		storage["deployments/status"] = deploymentStorage.Status
		storage["deployments/rollback"] = deploymentStorage.Rollback
		storage["deployments/scale"] = deploymentStorage.Scale
	}
	if apiResourceConfigSource.ResourceEnabled(version.WithResource("ingresses")) {
		ingressStorage, ingressStatusStorage := ingressstore.NewREST(restOptionsGetter, stopCh)
		storage["ingresses"] = ingressStorage
		storage["ingresses/status"] = ingressStatusStorage
	}
	if apiResourceConfigSource.ResourceEnabled(version.WithResource("podsecuritypolicy")) || apiResourceConfigSource.ResourceEnabled(version.WithResource("podsecuritypolicies")) {
		podSecurityExtensionsStorage := pspstore.NewREST(restOptionsGetter, stopCh)
		storage["podSecurityPolicies"] = podSecurityExtensionsStorage
	}
	if apiResourceConfigSource.ResourceEnabled(version.WithResource("replicasets")) {
		replicaSetStorage := replicasetstore.NewStorage(restOptionsGetter, stopCh)
		storage["replicasets"] = replicaSetStorage.ReplicaSet
		storage["replicasets/status"] = replicaSetStorage.Status
		storage["replicasets/scale"] = replicaSetStorage.Scale
	}
	if apiResourceConfigSource.ResourceEnabled(version.WithResource("networkpolicies")) {
		networkExtensionsStorage := networkpolicystore.NewREST(restOptionsGetter, stopCh)
		storage["networkpolicies"] = networkExtensionsStorage
	}

	return storage
}

func (p RESTStorageProvider) GroupName() string {
	return extensions.GroupName
}
