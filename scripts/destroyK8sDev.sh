#!/usr/bin/env bash

set -ex

# Delete Service Account
kubectl delete -f deploy/service_account.yaml

# Delete RBAC
kubectl delete -f deploy/role.yaml
kubectl delete -f deploy/role_binding.yaml

# Delete the CRD
kubectl delete -f deploy/crds/trs.hms.cray.com_trsworkers_crd.yaml