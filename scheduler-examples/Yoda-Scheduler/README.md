
## pkg

yoda-scheduler各个扩展点的代码逻辑：

* sort: 通过“scv/priority"标签比较两个pod调度的优先级
* filter（预选）: 通过PodFitsNumber, PodFitsMemory, PodFitsClock函数筛选出满足条件的节点
