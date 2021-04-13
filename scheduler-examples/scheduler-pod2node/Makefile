# Image URL to use all building/pushing image targets
IMG ?= haroldmua/scheduler-pod2node:v5

# Build the docker image
docker-build:
	docker build . -t ${IMG}

# Push the docker image
docker-push:
	docker push ${IMG}

scheduler-run:
	kubectl apply -f deploy/scheduler.yaml

scheduler-stop:
	kubectl delete -f deploy/scheduler.yaml

scheduler-log:
	kubectl logs scheduler

pod-run:
	kubectl apply -f deploy/pod.yaml

pod-stop:
	kubectl delete -f deploy/pod.yaml

