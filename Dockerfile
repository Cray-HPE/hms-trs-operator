# Copyright 2020 Hewlett Packard Enterprise Development LP

# Dockerfile for building HMS TRS Operator.

# Build base just has the packages installed we need.
FROM dtr.dev.cray.com/baseos/golang:1.14-alpine3.12 AS build-base

RUN set -ex \
    && apk update \
    && apk add --no-cache \
        build-base \
        git

# Base copies in the files we need to test/build.
FROM build-base AS base

WORKDIR /build

# Copy all the necessary files to the image.
COPY cmd     cmd
COPY pkg     pkg
COPY version version
COPY vendor  vendor

# Copy the Go module files.
COPY go.mod .
COPY go.sum .

### Build Stage ###
FROM base AS builder

ARG go_build_args="-mod=vendor"

RUN set -ex \
    && go build ${go_build_args} -v -o /usr/local/bin/hms-trs-operator ./cmd/manager

## Final Stage ###

FROM dtr.dev.cray.com/baseos/alpine:3.12
LABEL maintainer="Cray, Inc."

COPY --from=builder /usr/local/bin/hms-trs-operator /usr/local/bin

COPY .version /.version

RUN set -ex \
    && apk update \
    && apk add --no-cache curl

ENV WATCH_NAMESPACE=services
ENV TRS_IMAGE_PREFIX="dtr.dev.cray.com/"
ENV TRS_WORKER_LOG_LEVEL=INFO
ENV TRS_WORKER_KAFKA_BROKER_SPEC=cray-shared-kafka-kafka-bootstrap.services.svc.cluster.local:9092
ENV TRS_KAFKA_CLUSTER_NAME=cray-shared-kafka

CMD ["sh", "-c", "hms-trs-operator"]
