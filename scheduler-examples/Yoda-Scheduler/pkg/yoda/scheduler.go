package yoda

import (
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
	Plugin
	// Less are used to sort pods in the scheduling queue.
	Less(*QueuedPodInfo, *QueuedPodInfo) bool
}
QueueSortPlugin is an interface that must be implemented by "QueueSort" plugins.
These plugins are used to sort pods in the scheduling queue.
Only one queue sort plugin may be enabled at a time.

因此，为实现QueueSortPlugin接口，要实现Less方法
各接口对应的方法：
QueueSortPlugin： Less
FilterPlugin:     Filter
PostFilterPlugin: PostFilter
ScorePlugin:      Score, ScoreExtensions
ScoreExtensions:  NormalizeScore
 */
var (
	_ framework.QueueSortPlugin  = &Yoda{}
	_ framework.FilterPlugin     = &Yoda{}
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
func New(_ *runtime.Unknown, h framework.Handle) (framework.Plugin, error) { // https://github.com/kubernetes/kubernetes/blob/master/pkg/scheduler/framework/plugins/examples/multipoint/multipoint.go#L89

	// refer to: SCV/main.go
	if err := scv.AddToScheme(scheme); err == nil {
		klog.Error(err)
		return nil, err
	}

	// no idea what's this for
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
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

	return &Yoda{
		handle: h,
		cache: mgr.GetCache(),
	}, nil
}
func Filter()

func PostFilter()

func Less()

func Score()

func NormalizeScore()

func ScoreExtensions()