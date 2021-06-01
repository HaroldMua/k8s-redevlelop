package pkg

import (
	"context"
	"log"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

// Name is the name of the plugin used in the plugin registry and configurations.
const Name = "sample2"

type sample struct{}

var _ framework.FilterPlugin = &sample{}
var _ framework.PreScorePlugin = &sample{}

// Name returns name of the plugin.
func (pl *sample) Name() string {
	return Name
}
func (pl *sample) Filter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	log.Printf("Print info. filter pod: %v, node: %v", pod.Name, nodeInfo)
	log.Println(state)

	// 排除没有cpu=true标签的节点
	if nodeInfo.Node().Labels["cpu"] != "true" {
		return framework.NewStatus(framework.Unschedulable, "Print info. Node: "+nodeInfo.Node().Name)
	}
	return framework.NewStatus(framework.Success, "Node: "+nodeInfo.Node().Name)
}

func (pl *sample) PreScore(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodes []*v1.Node) *framework.Status  {
	log.Println(nodes)
	return framework.NewStatus(framework.Success, "Node: "+pod.Name)
}

// New initializes a new plugin and returns it.
func New(_ runtime.Object, _ framework.Handle) (framework.Plugin, error) {
	return &sample{}, nil
}