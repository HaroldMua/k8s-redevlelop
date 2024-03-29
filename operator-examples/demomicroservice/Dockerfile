# Build the manager binary
FROM golang:1.13 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

#-------
# set GOPROXY to avoid "time out" errors on "go get pkg".
ENV GOPROXY=https://goproxy.cn,direct
#--------

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/
# util is self-built
COPY util/ util/


# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o manager main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details

# Because of the network, use "exploitht/operator-static:latest" to replace "gcr.io/distroless/static:nonroot".
#FROM gcr.io/distroless/static:nonroot
FROM  exploitht/operator-static:latest
WORKDIR /
COPY --from=builder /workspace/manager .
USER nonroot:nonroot

ENTRYPOINT ["/manager"]
