FROM golang:1.15-alpine AS builder
WORKDIR /go/src/github.com/woblerr/prom_authlog_exporter
COPY . .
RUN apk update \
    && apk add git \
    && go get -d -v ./... \
    && CGO_ENABLED=0 GOOS=linux go build -o auth_exporter auth_exporter.go


FROM scratch
COPY --from=builder /go/src/github.com/woblerr/prom_authlog_exporter/auth_exporter /auth_exporter
EXPOSE 9991
USER nobody
ENTRYPOINT ["/auth_exporter"]
CMD ["-h"]
