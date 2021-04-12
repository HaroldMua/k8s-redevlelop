FROM golang:1.14-alpine as builder
ADD . /go/src/github.com/HaroldMua/k8s-redevelop/devicePlugin-examples/k8s-device-plugin-example
WORKDIR /go/src/github.com/HaroldMua/k8s-redevelop/devicePlugin-examples/k8s-device-plugin-example/cmd

# set GOPROXY to avoid "time out" errors on "go get pkg".
ENV GOPROXY=https://goproxy.cn,direct
RUN go build -v
 
FROM alpine:latest
COPY --from=builder /go/src/github.com/HaroldMua/k8s-redevelop/devicePlugin-examples/k8s-device-plugin-example/cmd/cmd /usr/local/bin/k8s-device-plugin-example
CMD ["k8s-device-plugin-example"]