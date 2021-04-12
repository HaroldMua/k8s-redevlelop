# Image URL to use all building/pushing image targets
IMG ?= haroldmua/k8s-device-plugin-example:v1

# Build the docker image
docker-build:
	docker build . -t ${IMG}

# Push the docker image
docker-push:
	docker push ${IMG}