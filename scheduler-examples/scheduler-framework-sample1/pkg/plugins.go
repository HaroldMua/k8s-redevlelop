package pkg

import (
	"context"
	"scheduler-framework-sample1/pkg/score"
	"strconv"
	"strings"

	"scheduler-framework-sample1/pkg/filter"
	//"scheduler-framework-sample1/pkg/prefilter"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

// Name is the name of the plugin used in the plugin registry and configurations.
const Name = "sample"

// Sort is a plugin that implements QoS class based sorting.
type sample struct{}

/*
各接口对应的方法：
PreFilterPlugin:  PreFilter, PreFilterExtensions
FilterPlugin:     Filter
ScorePlugin:      Score, ScoreExtensions
*/
var (
	_ framework.PreFilterPlugin = &sample{}
	_ framework.FilterPlugin    = &sample{}
	_ framework.ScorePlugin     = &sample{}
	_ framework.ScoreExtensions = &sample{}
)

// Name returns name of the plugin.
func (pl *sample) Name() string {
	return Name
}

/*
由于"rely.on.pod/labs"依赖的pod无法创建，故注释掉PreFilterPlugin的代码
 */
func (pl *sample) PreFilter(ctx context.Context, state *framework.CycleState, p *v1.Pod) *framework.Status {

	/*
		通过pod已声明的Annotations参数 rely.on.namespaces/name 和 rely.on.pod/labs来获取该pod依赖的pod是否就绪，
		如依赖的pod未就绪，则终止调度，pod处于Pending状态。
	*/
	//namespace := p.Annotations["rely.on.namespaces/name"]
	//podLabs := p.Annotations["rely.on.pod/labs"]
	//
	//if namespace == "" || podLabs == "" {
	//	return framework.NewStatus(framework.Success, "ont rely")
	//}
	//if prefilter.IsExist(namespace) == false {
	//	return framework.NewStatus(framework.Unschedulable, "not found namespace: "+namespace)
	//}
	//
	//if prefilter.IsReady(namespace, podLabs) == false {
	//	return framework.NewStatus(framework.Unschedulable, "rely pod not ready")
	//}
	//klog.Infoln("rely pod is ready :", namespace, podLabs, prefilter.IsReady(namespace, podLabs))
	return framework.NewStatus(framework.Success, "rely pod is ready")
}

func (pl *sample) PreFilterExtensions() framework.PreFilterExtensions {
	return nil
}


func (pl *sample) Filter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, node *framework.NodeInfo) *framework.Status {

	/*
	过滤掉没有 cpu=true labels的节点；
	Pod调度资源值和限制资源值大于节点当前可用资源的节点，则过滤；
	默认过滤cpu内存使用率超过85%的节点。CPU_THRESHOLD 和  MEN_THRESHOLD 环境变量设置该值。
	 */

	if node.Node().Labels["cpu"] != "true" {
		return framework.NewStatus(framework.Unschedulable, "not found labels")
	}
	nodeUsedCPU, nodeUsedMen, nodeCPU, nodeMen, cpuRate, menRate := filter.ResourceStatus(node.Node().Name)

	for i := 0; i < len(pod.Spec.Containers); i++ {
		requestsCPUCore, _ := strconv.ParseFloat(strings.Replace(pod.Spec.Containers[i].Resources.Requests.Cpu().String(), "n", "", 1), 64)
		requestsCPU := requestsCPUCore * 1000 * (1000 * 1000)
		requestsMen := pod.Spec.Containers[i].Resources.Requests.Memory().Value() / 1024 / 1024
		limitsCPUCore, _ := strconv.ParseFloat(strings.Replace(pod.Spec.Containers[i].Resources.Limits.Cpu().String(), "n", "", 1), 64)
		limitsCPU := limitsCPUCore * 1000 * (1000 * 1000)
		limitsMen := pod.Spec.Containers[i].Resources.Limits.Memory().Value() / 1024 / 1024
		if requestsCPU > float64(nodeCPU) || requestsMen > nodeMen {
			return framework.NewStatus(framework.Unschedulable, "out of Requests resources")
		}
		if limitsCPU > float64(nodeCPU) || limitsMen > nodeMen {
			return framework.NewStatus(framework.Unschedulable, "out of Limits resources")
		}
		if requestsCPU > float64(nodeCPU)-nodeUsedCPU || requestsMen > (nodeMen-nodeUsedMen) {
			return framework.NewStatus(framework.Unschedulable, "out of Requests resources system")
		}
		if limitsCPU > float64(nodeCPU)-nodeUsedCPU || limitsMen > (nodeMen-nodeUsedMen) {
			return framework.NewStatus(framework.Unschedulable, "out of Limits resources system")
		}

	}

	klog.Infof("node:%s, CPU:%v ,  Memory: %v", node.Node().Name, cpuRate, menRate)
	cpuThreshold := filter.GetEnvFloat("CPU_THRESHOLD", 0.85)
	menThreshold := filter.GetEnvFloat("MEN_THRESHOLD", 0.85)
	if cpuRate > cpuThreshold || menRate > menThreshold {
		return framework.NewStatus(framework.Unschedulable, "out of system resources")
	}

	return framework.NewStatus(framework.Success, "Node: "+node.Node().Name)
}

func (pl *sample) Score(ctx context.Context, state *framework.CycleState, p *v1.Pod, nodeName string) (int64, *framework.Status) {

	/*
	打分规则：
	配置高的节点权重大
	当前资源使用率底的节点权重大
	运行多组pod的情况下，运行相同pod的节点权重低
	 */
	isSamePod := score.IsSamePod(nodeName, p.Namespace, p.Labels) // max 2
	cpuLoad := score.CPURate(nodeName)                            // max 3
	menLoad := score.MemoryRate(nodeName)                         // max 3
	core := score.CpuCore(nodeName)                               // max 3

	c := isSamePod + cpuLoad + core + menLoad
	klog.Infoln(nodeName+" score is :", c)
	return c, framework.NewStatus(framework.Success, nodeName)
}

func (pl *sample) NormalizeScore(ctx context.Context, state *framework.CycleState, p *v1.Pod, scores framework.NodeScoreList) *framework.Status {
	var (
		highest int64 = 0
		lowest        = scores[0].Score
	)
	klog.Infoln("--------->", scores)
	for _, nodeScore := range scores {
		klog.Infoln("highest for:--------->", highest)
		klog.Infoln("lowest for:--------->", lowest)
		if nodeScore.Score < lowest {
			lowest = nodeScore.Score
		}
		if nodeScore.Score > highest {
			highest = nodeScore.Score
		}
	}
	klog.Infoln("highest:--------->", highest)
	klog.Infoln("lowest:--------->", lowest)
	if highest == lowest {
		lowest--
	}

	for i, nodeScore := range scores {
		scores[i].Score = (nodeScore.Score - lowest) * framework.MaxNodeScore / (highest - lowest)
		klog.Infof("node: %v, final Score: %v", scores[i].Name, scores[i].Score)
	}
	return framework.NewStatus(framework.Success, "")
}

func (pl *sample) ScoreExtensions() framework.ScoreExtensions {
	return nil
}

// New initializes a new plugin and returns it.
func New(_ runtime.Object, _ framework.Handle) (framework.Plugin, error) {
	return &sample{}, nil
}
