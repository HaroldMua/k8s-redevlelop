IMG ?= haroldmua/scheduler-framework-sample1:v2

all: build

build:
		CGO_ENABLED=0 GOOS=linux go build -o app

# Build the docker image
docker-build:
		docker build . -t ${IMG}

# Push the docker image
docker-push:
		docker push ${IMG}