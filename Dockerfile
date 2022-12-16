FROM golang:1.19 as basebuilder

RUN set -x \
    && apt-get update \
    && apt-get install -y make jq

FROM basebuilder as builder
ENV GOGC off
ENV CGO_ENABLED 0
ARG VERSION=dev
ARG REPO=repository

WORKDIR /src
COPY . /src

RUN make

# Executable image
FROM scratch

WORKDIR /

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /src/bin/frostfs-rest-gw /bin/frostfs-rest-gw
COPY --from=builder /src/static /static

ENTRYPOINT ["/bin/frostfs-rest-gw"]
