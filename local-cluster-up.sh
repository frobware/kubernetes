#!/bin/bash

export KUBERNETES_PROVIDER=local
export API_HOST=`ifconfig docker0 | grep -w "inet" | awk -F'[: ]+' '{ print $3 }'`
export KUBE_ENABLE_CLUSTER_DNS=true
export KUBELET_HOST=0.0.0.0
export HOSTNAME_OVERRIDE=`ifconfig enp5s0 | grep -w "inet" | awk -F'[: ]+' '{ print $3 }'`

./hack/local-up-cluster.sh
