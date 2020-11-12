#!/usr/bin/env bash

set -ex

# Setup Service Account
kubectl apply -f deploy/service_account.yaml

# Setup RBAC
kubectl apply -f deploy/role.yaml
kubectl apply -f deploy/role_binding.yaml

# Setup the CRDs
kubectl apply -f deploy/crds/trs.hms.cray.com_trsworkers_crd.yaml
#kubectl apply -f deploy/crds/kafka.strimzi.io_kafkatopics_crd.yaml