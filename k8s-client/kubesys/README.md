# [kubernetes-client-go](https://github.com/kubesys/kubernetes-client-go)

### Create a client

```
client := kubesys.NewKubernetesClient(url, tok)
client.Init()
```

Here, the token can be created and get by following commands:

1. create token
```
kubectl apply -f account.yaml
```

2. get token
```
kubectl -n kube-system describe secret $(kubectl -n kube-system get secret | grep kubernetes-client | awk '{print $1}') | grep "token:" | awk -F":" '{print$2}' | sed 's/ //g'
```