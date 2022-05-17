# HMS TRS Operator

## Overview

This repo contains all the pieces to the operator for the TRS project. The lifecycle of this operator is to make sure worker pods are available for handling the work that needs to be done by client applications.

More info available at: https://github.com/Cray-HPE/hms-trs-app-api

## TRS Operator Manual Smoke Testing

1. Install the new cray-hms-trs-operator
2. Create a TRSWorker using the config file [testapp1-http-v1.yaml](deploy/crds/test_crs/testapp1-http-v1.yaml)

```bash
kubectl apply -n services -f testapp1-http-v1.yaml
kubectl -n services get TRSWorker
```

3. Check that the TRS pods were created.

```bash
kubectl -n services get pods | grep trs
```

4. Check the logs of cray-hms-trs-operator

```bash
kubectl -n operators get pods | grep -i trs
kubectl -n operators logs cray-hms-trs-operator-697f87c8b-dpwvn
```

5. When done remove the TRSWorker

```bash
kubectl delete -n services -f testapp1-http-v1.yaml
```
