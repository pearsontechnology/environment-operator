# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang:1.10-alpine AS builder

ADD https://github.com/golang/dep/releases/download/v0.5.0/dep-linux-amd64 /usr/bin/dep
RUN chmod +x /usr/bin/dep

RUN apk update && \
    apk add git

# Copy the local package files to the container's workspace.
COPY . /go/src/github.com/pearsontechnology/environment-operator
WORKDIR /go/src/github.com/pearsontechnology/environment-operator

# Only install dependencies specified in dep vendor file
RUN dep ensure --vendor-only
COPY . ./

RUN adduser -u 1000 -S oper
USER oper

RUN CGO_ENABLED=1 go test -v ./pkg/bitesize ./pkg/cluster ./pkg/diff ./pkg/git ./pkg/reaper ./pkg/translator ./pkg/web ./pkg/util ./pkg/util/k8s

RUN go build -v -o /tmp/environment-operator ./cmd/operator/main.go

# Second stage
FROM alpine:3.8
RUN apk update && apk add curl
COPY --from=builder /tmp/environment-operator /environment-operator
ENTRYPOINT ["/environment-operator"]
