# MIT License

# (C) Copyright [2021] Hewlett Packard Enterprise Development LP

# Permission is hereby granted, free of charge, to any person obtaining a
# copy of this software and associated documentation files (the "Software"),
# to deal in the Software without restriction, including without limitation
# the rights to use, copy, modify, merge, publish, distribute, sublicense,
# and/or sell copies of the Software, and to permit persons to whom the
# Software is furnished to do so, subject to the following conditions:
# The above copyright notice and this permission notice shall be included
# in all copies or substantial portions of the Software.

# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
# THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
# OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
# ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
# OTHER DEALINGS IN THE SOFTWARE.

# Dockerfile for building HMS TRS Operator.

# Build base just has the packages installed we need.
FROM arti.dev.cray.com/baseos-docker-master-local/golang:1.16-alpine3.13 AS build-base

RUN set -ex \
    && apk update \
    && apk add --no-cache \
        build-base \
        git

# Base copies in the files we need to test/build.
FROM build-base AS base

WORKDIR /build

RUN go env -w GO111MODULE=auto

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

FROM arti.dev.cray.com/baseos-docker-master-local/alpine:3.13
LABEL maintainer="Hewlett Packard Enterprise"

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
