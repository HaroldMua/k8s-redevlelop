package sort

import (
	"k8s.io/kubernetes/pkg/scheduler/framework"
	"strconv"
)

func Less(podInfo1, podInfo2 *framework.QueuedPodInfo) bool {
	return GetPodPriority(podInfo1) > GetPodPriority(podInfo2)
}

// QueuedPodInfo is a Pod wrapper with additional information related to
// the pod's status in the scheduling queue, such as the timestamp when it's added to the queue.
func GetPodPriority(podInfo *framework.QueuedPodInfo) int {
	if p, ok := podInfo.Pod.GetLabels()["scv/priority"]; ok {
	//if p, ok := podInfo.Pod.Labels["scv/priority"]; ok {   // 这是另外一种写法？
		pri, _ := strconv.Atoi(p)
		return pri
	}
	return 0
}