ARG REPO_BUILD_TAG="unknown"

FROM golang:1.16-alpine AS builder
ARG REPO_BUILD_TAG
WORKDIR /go/src/github.com/woblerr/prom_authlog_exporter
COPY . .
RUN apk update \
    && apk add git \
    && CGO_ENABLED=0 go build \
        -mod=vendor -trimpath \
        -ldflags "-X main.version=${REPO_BUILD_TAG}" \
        -o auth_exporter auth_exporter.go

FROM scratch
ARG REPO_BUILD_TAG
COPY --from=builder /go/src/github.com/woblerr/prom_authlog_exporter/auth_exporter /auth_exporter
EXPOSE 9991
USER nobody
ENTRYPOINT ["/auth_exporter"]
CMD ["-h"]
