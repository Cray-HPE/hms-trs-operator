#!/usr/bin/env bash

set -x

kubectl create namespace services
kubectl create namespace operators

helm upgrade --install --namespace services strimzi strimzi/strimzi-kafka-operator
kubectl -n services apply -f deploy/strimzi_kafka.yaml

helm upgrade --install --namespace operators \
  --set imagesHost='localhost:5000' \
  --set workerLogLevel='TRACE' \
  --set workerKafkaBrokerSpec='k8s-kafka-kafka-bootstrap:9092' \
  --set workerKafkaClusterName='k8s-kafka' \
  trs kubernetes/cray-hms-trs-operator