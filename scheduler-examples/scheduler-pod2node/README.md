## Kubernetes Scheduler Example

This repository contains a Kubernetes scheduler code for random node assignment. Refer to [k8s-scheduler-example](https://github.com/onuryilmaz/k8s-scheduler-example)

### Build and push
```
$ make docker-build
$ make docker-push
```


### Example usage


#### Create a pod with custom scheduler:
```
$ kubectl apply -f deploy/pod.yaml
$ kubectl get pods
NAME  		READY     STATUS    RESTARTS   AGE
nginx          0/1       Pending   0          5s
```

#### Deploy the scheduler into cluster:
```
$ kubectl apply -f deploy/scheduler.yaml
```

#### Check the pods:
```
$ kubectl get pods
NAME        READY     STATUS    RESTARTS   AGE
nginx       1/1       Running   0          44s
scheduler   1/1       Running   0          17s
```

#### Check the logs of scheduler:
```
$ kubectl logs scheduler
Starting scheduler: packt-scheduler
Assigning default/nginx to minikube 
```

### Cleanup
```
$ kubectl delete -f deploy/pod.yaml
$ kubectl delete -f deploy/scheduler.yaml
```
