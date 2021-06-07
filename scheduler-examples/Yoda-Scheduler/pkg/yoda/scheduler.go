package yoda

import (
	"context"
	"errors"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"

	scv "SCV/api/v1"

	"Yoda-Scheduler/pkg/yoda/collection"
	"Yoda-Scheduler/pkg/yoda/filter"
	"Yoda-Scheduler/pkg/yoda/score"
	"Yoda-Scheduler/pkg/yoda/sort"
)

const Name = "yoda"

type Yoda struct {
	handle framework.Handle
	cache cache.Cache
}

/*
refer to: https://pkg.go.dev/k8s.io/kubernetes/pkg/scheduler/framework#QueueSortPlugin

type QueueSortPlugin interface {
    // 嵌套Plugin接口
	Plugin
	// Less are used to sort pods in the scheduling queue.
	Less(*QueuedPodInfo, *QueuedPodInfo) bool
}
QueueSortPlugin is an interface that must be implemented by "QueueSort" plugins.
These plugins are used to sort pods in the scheduling queue.
Only one queue sort plugin may be enabled at a time.

因此，结构体要实现对应的接口方法，如为实现QueueSortPlugin接口，要实现Less方法
各接口对应的方法：
QueueSortPlugin： Less
FilterPlugin:     Filter
PostFilterPlugin: PostFilter
ScorePlugin:      Score, ScoreExtensions
ScoreExtensions:  NormalizeScore
 */
var (
	_ framework.QueueSortPlugin  = &Yoda{}   // 确保指针接收者Yoda实现了接口framework.QueueSortPlugin
	_ framework.FilterPlugin     = &Yoda{}   // 为了保护你的Go语言职业生涯，请牢记接口是一种类型
	_ framework.PostFilterPlugin = &Yoda{}
	_ framework.ScorePlugin      = &Yoda{}
	_ framework.ScoreExtensions  = &Yoda{}

	scheme = runtime.NewScheme()
)

func (y *Yoda) Name() string {
	return Name
}

// New initializes a new plugin and returns it.
// framework.Plugin is the parent type for all the scheduling framework plugins.
func New(_ runtime.Object, h framework.Handle) (framework.Plugin, error) {
	mgrConfig := ctrl.GetConfigOrDie()
	mgrConfig.QPS = 1000
	mgrConfig.Burst = 1000

	if err := scv.AddToScheme(scheme); err != nil {
		klog.Error(err)
		return nil, err
	}

	mgr, err := ctrl.NewManager(mgrConfig, ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: "",
		LeaderElection:     false,
		Port:               9443,
	})
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	go func() {
		if err = mgr.Start(ctrl.SetupSignalHandler()); err != nil {
			klog.Error(err)
			panic(err)
		}
	}()

	scvCache := mgr.GetCache()

	if scvCache.WaitForCacheSync(context.TODO()) {
		return &Yoda{
			handle: h,
			cache:  scvCache,
		}, nil
	} else {
		return nil, errors.New("Cache Not Sync! ")
	}
}

func (y *Yoda) Filter(ctx context.Context, _ *framework.CycleState, pod *v1.Pod, node *framework.NodeInfo) *framework.Status {
	klog.V(3).Infof("filter pod: %v, node: %v", pod.Name, node.Node().Name)

	currentScv := &scv.Scv{}


	/*
	ClientSet仅能访问k8s自身内置的资源（即客户端集合内的资源），
	不能直接访问CRD自定义资源。如果需要ClientSet访问CRD自定义资源，
	可通过client-gen代码生成器重新生成ClientSet,在ClientSet集合中自动生成与CRD操作相关的接口

	Refer to: https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/cache#Cache
	Cache knows how to load Kubernetes objects, fetch informers to request to receive events for
	Kubernetes objects (at a low-level), and add indices to fields on the objects stored in the cache.
	 */
	// Get Scv
	err := y.cache.Get(ctx, types.NamespacedName{Name: node.Node().GetName()}, currentScv)
	if err != nil {
		klog.Errorf("Get SCV Error: %v", err)
		return framework.NewStatus(framework.Unschedulable, "Node:"+node.Node().Name+" "+err.Error())  //The error built-in interface type have Error() method
	}

	// Alright, this is added filter policy
	if ok, number := filter.PodFitsNumber(pod, currentScv); ok {
		isFitsMemory, _ := filter.PodFitsMemory(number, pod, currentScv)
		isFitsClock, _ := filter.PodFitsClock(number, pod, currentScv)
		if isFitsMemory && isFitsClock {
			return framework.NewStatus(framework.Success, "")
		}
	}

	return framework.NewStatus(framework.Unschedulable, "Node:"+node.Node().Name)
}

func (y *Yoda) PostFilter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, _ framework.NodeToStatusMap) (*framework.PostFilterResult, *framework.Status) {
	klog.V(3).Infof("collect info for scheduling pod: %v", pod.Name)
	scvList := scv.ScvList{}

	if err := y.cache.List(ctx, &scvList); err != nil {
		klog.Errorf("Get Scv List Error: %v", err)
		return &framework.PostFilterResult{}, framework.NewStatus(framework.Error, err.Error())
	}

	return &framework.PostFilterResult{}, collection.CollectMaxValues(state, pod, scvList)   // why return collection.CollectMaxValues
}

func (y *Yoda) Less(podInfo1, podInfo2 *framework.QueuedPodInfo) bool {
	return sort.Less(podInfo1, podInfo2)
}

func (y *Yoda) Score(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {
	// refer to: https://github.com/kubernetes/kubernetes/blob/master/pkg/scheduler/framework/plugins/nodelabel/node_label.go#L117
	// Get Node Info
	nodeInfo, err := y.handle.SnapshotSharedLister().NodeInfos().Get(nodeName)
	if err != nil {
		return 0, framework.AsStatus(fmt.Errorf("getting node %q from Snapshot: %w", nodeName, err))
	}

	// Get Scv info
	currentScv := &scv.Scv{}
	err = y.cache.Get(ctx, types.NamespacedName{Name: nodeName}, currentScv)
	if err != nil {
		klog.Errorf("Get SCV Error: %v", err)
		return 0, framework.NewStatus(framework.Error, fmt.Sprintf("Score Node Error: %v", err))
	}

	uNodeScore, err := score.CalculateScore(currentScv, state, pod, nodeInfo)
	if err != nil {
		return 0, framework.NewStatus(framework.Error, fmt.Sprintf("Score Node Error: %v", err))
	}
	nodeScore := filter.Uint64ToInt64(uNodeScore)
	return nodeScore, framework.NewStatus(framework.Success, "")
}

// NormalizeScore invoked after scoring all nodes.
func (y *Yoda) NormalizeScore(_ context.Context, _ *framework.CycleState, pod *v1.Pod, scores framework.NodeScoreList) *framework.Status {
	var (
		highest int64 = 0
		lowest        = scores[0].Score
	)

	for _, nodeScore := range scores {
		if nodeScore.Score < lowest {
			lowest = nodeScore.Score
		}
		if nodeScore.Score > highest {
			highest = nodeScore.Score
		}
	}

	if highest == lowest {
		lowest--
	}

	// Set Range to [0-100]
	for i, nodeScore := range scores {
		scores[i].Score = (nodeScore.Score - lowest) * framework.MaxNodeScore / (highest - lowest)
		klog.Infof("Node: %v, Score: %v in Plugin: Mandalorian When scheduling Pod: %v/%v", scores[i].Name, scores[i].Score, pod.GetNamespace(), pod.GetName())
	}
	return nil
}

func (y *Yoda) ScoreExtensions() framework.ScoreExtensions {
	return y
}