# Copyright 2020 Hewlett Packard Enterprise Development LP

# This image is for development purposes only and is not
# used as part of any deliverable artifact from HMS.
FROM registry.access.redhat.com/ubi8/ubi-minimal:latest

ENV OPERATOR=/usr/local/bin/trs-operator \
    USER_UID=1001 \
    USER_NAME=trs-operator

# install operator binary
COPY build/_output/bin/hms-trs-operator ${OPERATOR}

COPY build/bin /usr/local/bin
RUN  /usr/local/bin/user_setup

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}
