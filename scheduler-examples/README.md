
宏观来看，调度器就是利用API服务器的监听机制（结合informer机制，通过informer机制可以很容易监控我们所关心的资源事件）等待新创建的Pod,然后给每个新的，没有节点集的pod分配节点。

简而言之，调度器就是给pod分配节点。

![scheduler](../doc/scheduler.png)