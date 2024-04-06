ARG REPO_BUILD_TAG="unknown"

FROM golang:1.21-alpine3.19 AS builder
ARG REPO_BUILD_TAG
WORKDIR /go/src/github.com/woblerr/authlog_exporter
COPY . .
RUN apk update \
    && apk add git \
    && CGO_ENABLED=0 go build \
        -mod=vendor -trimpath \
        -ldflags "-X main.version=${REPO_BUILD_TAG}" \
        -o authlog_exporter authlog_exporter.go

FROM alpine:3.19
ARG REPO_BUILD_TAG
RUN apk add --no-cache --update ca-certificates \
    && rm -rf /var/cache/apk/*
COPY --from=builder /go/src/github.com/woblerr/authlog_exporter/authlog_exporter /authlog_exporter
EXPOSE 9991
USER nobody
LABEL \
    org.opencontainers.image.version="${REPO_BUILD_TAG}" \
    org.opencontainers.image.source="https://github.com/woblerr/authlog_exporter"
ENTRYPOINT ["/authlog_exporter"]
CMD ["-h"]
