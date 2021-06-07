# Image URL to use all building/pushing image targets
IMG ?= haroldmua/yoda-scheduler:v6

all: local

local: fmt vet
	GOOS=linux GOARCH=amd64 go build  -o=bin/yoda-scheduler ./cmd/scheduler

build:
	docker build --no-cache . -t ${IMG}

push:
	docker push ${IMG}

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

clean: fmt vet
	sudo rm -f yoda-scheduler