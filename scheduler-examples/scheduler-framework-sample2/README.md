## Scheduler Plugins Sample

参照社区示例的[QoS插件](https://github.com/kubernetes-sigs/scheduler-plugins/blob/master/pkg/qos/queue_sort.go), 简单开发一个Filter, PreScore的插件。

该示例基于[scheduler framework](https://kubernetes.io/zh/docs/concepts/scheduling-eviction/scheduling-framework/) 搭建了开发插件的基本框架，缺乏调度逻辑。参考[这里](https://mp.weixin.qq.com/s/NGWSv0iF2_cwKJt7AdLXxQ).

### 调度器配置

根据[官方文档](https://kubernetes.io/zh/docs/reference/scheduling/config/), 可以通过编写配置文件，并将其路径传给 kube-scheduler 的命令行参数，定制 kube-scheduler 的行为。
调度模板（Profile）允许你配置 kube-scheduler 中的不同调度阶段。每个阶段都暴露于某个扩展点中。插件通过实现一个或多个扩展点来提供调度行为。

你可以通过运行 kube-scheduler --config <filename> 来设置调度模板， 使用 [KubeSchedulerConfiguration (v1beta1)](https://kubernetes.io/docs/reference/config-api/kube-scheduler-config.v1beta1/#kubescheduler-config-k8s-io-v1beta1-KubeSchedulerConfiguration) 结构体。


[./sample-scheduler.yaml](./sample-scheduler.yaml)需要注意的几点:

```
apiVersion: v1
kind: ConfigMap
metadata:
  name: scheduler-config
  namespace: kube-system
data:
  # refer to: https://kubernetes.io/zh/docs/reference/scheduling/config/
  scheduler-config.yaml: |
    apiVersion: kubescheduler.config.k8s.io/v1beta1
    kind: KubeSchedulerConfiguration
    leaderElection:
      leaderElect: false
    profiles:
    - schedulerName: sample-scheduler
      plugins:
        filter:
          enabled:
          - name: sample2 # FilterPlugin名称与pkg/plugins.go定义的plugin名称一致
        preScore:
          enabled:
            - name: sample2
          disabled:
            - name: "*"
```

```sample-scheduler.yaml
      ...
      containers:
        - name: scheduler-framework
          image: haroldmua/scheduler-framework:v1
          imagePullPolicy: IfNotPresent
           # docker:ENTRYPOINT = k8s:command
           # docker:CMD = k8s:args
          command:   # 传给kube-scheduler的命令行参数
            - app   # Dockerfile中编译项目生成的二进制执行文件名为"app"
            - --config=/etc/kubernetes/scheduler-config.yaml   # 挂载的configmap配置文件参数   
            - --v=3
          ...
```

### 验证插件

使用[./test-scheduler.yaml](./test-scheduler.yaml)验证插件，schedulerName字段指定自定义调度器.

运行`kubectl apply -f ./test-scheduler.yaml`后，查看pod的描述信息：

```
Name:         test-scheduler-cb8bdd788-gm6pk
Namespace:    default
Priority:     0
Node:         gpu02-poweredge-t420/192.168.1.16
Start Time:   Tue, 01 Jun 2021 15:14:11 +0800
Labels:       app=test-scheduler
              pod-template-hash=cb8bdd788
Annotations:  <none>
Status:       Running
IP:           10.244.2.85
IPs:
  IP:           10.244.2.85
Controlled By:  ReplicaSet/test-scheduler-cb8bdd788
Containers:
  nginx:
    Container ID:   docker://78c9a59946c0fae74daf1460f83139603f228c9065a251e9247497725af4b81b
    Image:          nginx:1.19.2-alpine
    Image ID:       docker-pullable://nginx@sha256:a97eb9ecc708c8aa715ccfb5e9338f5456e4b65575daf304f108301f3b497314
    Port:           80/TCP
    Host Port:      0/TCP
    State:          Running
      Started:      Tue, 01 Jun 2021 15:14:15 +0800
    Ready:          True
    Restart Count:  0
    Environment:    <none>
    Mounts:
      /var/run/secrets/kubernetes.io/serviceaccount from default-token-7zs8q (ro)
Conditions:
  Type              Status
  Initialized       True 
  Ready             True 
  ContainersReady   True 
  PodScheduled      True 
Volumes:
  default-token-7zs8q:
    Type:        Secret (a volume populated by a Secret)
    SecretName:  default-token-7zs8q
    Optional:    false
QoS Class:       BestEffort
Node-Selectors:  <none>
Tolerations:     node.kubernetes.io/not-ready:NoExecute for 300s
                 node.kubernetes.io/unreachable:NoExecute for 300s
Events:
  Type     Reason            Age                From                           Message
  ----     ------            ----               ----                           -------
  Warning  FailedScheduling  84s (x19 over 4m)  sample-scheduler               0/2 nodes are available: 1 Print info. Node: gpu02-poweredge-t420, 1 Print info. Node: gpu03-poweredge-t420.
  Normal   Scheduled         74s                sample-scheduler               Successfully assigned default/test-scheduler-cb8bdd788-gm6pk to gpu02-poweredge-t420
  Normal   Pulled            71s                kubelet, gpu02-poweredge-t420  Container image "nginx:1.19.2-alpine" already present on machine
  Normal   Created           70s                kubelet, gpu02-poweredge-t420  Created container nginx
  Normal   Started           70s                kubelet, gpu02-poweredge-t420  Started container nginx
```

因为排除没有 cpu=true 标签的节点，故pod调度失败，处于Pending状态，自定义插件日志也会打印信息

```
I0601 07:12:51.806374       1 factory.go:321] "Unable to schedule pod; no fit; waiting" pod="default/test-scheduler-cb8bdd788-gm6pk" err="0/2 nodes are available: 1 Print info. Node: gpu02-poweredge-t420, 1 Print info. Node: gpu03-poweredge-t420."```
```

给节点增加`cpu=true`标签后，Pod已经正常调度，处于running状态，同时自定义插件也输出对应的日志

```
I0601 07:14:11.822855       1 default_binder.go:51] Attempting to bind default/test-scheduler-cb8bdd788-gm6pk to gpu02-poweredge-t420
I0601 07:14:11.829111       1 scheduler.go:592] "Successfully bound pod to node" pod="default/test-scheduler-cb8bdd788-gm6pk" node="gpu02-poweredge-t420" evaluatedNodes=2 feasibleNodes=1
I0601 07:14:11.829175       1 eventhandlers.go:209] delete event for unscheduled pod default/test-scheduler-cb8bdd788-gm6pk
I0601 07:14:11.829234       1 eventhandlers.go:229] add event for scheduled pod default/test-scheduler-cb8bdd788-gm6pk 
```