## Scheduler Plugins Sample

基于[scheduler-framework-sample2](https://github.com/HaroldMua/k8s-redevlelop/tree/master/scheduler-examples/scheduler-framework-sample2) 的[scheduler framework](https://kubernetes.io/zh/docs/concepts/scheduling-eviction/scheduling-framework/) 插件开发框架，实现一个简单的基于CPU指标的自定义调度器。自定义调度器通过kubernetes资源指标服务metrics-server来获取各节点的当前的资源情况，并进行打分，然后把Pod调度到分数最高的节点。参考[这里](https://mp.weixin.qq.com/s/Ep0DOSE5Sf7yFAvxHmU8xw).

可以利用annotations信息来做PreFilter的判断：

```test-scheduler.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-scheduler
spec:
  replicas: 1
  selector:
    matchLabels:
      app: test-scheduler
  template:
    metadata:
      labels:
        app: test-scheduler
      # annotations信息用于PreFilter
      annotations:
        rely.on.namespaces/name: "kube-system"
        rely.on.pod/labs: "k8s-app=metrics-server"
    spec:
      schedulerName: sample-scheduler
      containers:
        - image: nginx:1.19.2-alpine
          imagePullPolicy: IfNotPresent
          name: nginx
          ports:
            - containerPort: 80
          resources:
            requests:
              cpu: 1000m
              memory: 1024Mi
            limits:
              cpu: 2000m
              memory: 2048Mi
```



