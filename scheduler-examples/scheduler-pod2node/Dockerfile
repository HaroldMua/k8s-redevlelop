FROM golang:1.14-alpine as builder
ADD . /go/src/github.com/HaroldMua/k8s-redevelop/scheduler-examples/scheduler-pod2node
WORKDIR /go/src/github.com/HaroldMua/k8s-redevelop/scheduler-examples/scheduler-pod2node

# set GOPROXY to avoid "time out" errors on "go get pkg".
ENV GOPROXY=https://goproxy.cn,direct
RUN go build -v -o scheduler

FROM alpine:latest
COPY --from=builder /go/src/github.com/HaroldMua/k8s-redevelop/scheduler-examples/scheduler-pod2node/scheduler /usr/local/bin/scheduler
CMD ["scheduler"]