package collection

import (
	scv "SCV/api/v1"
	"Yoda-Scheduler/pkg/yoda/filter"
	v1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

type Data struct {
	Value MaxValue
}

type MaxValue struct {
	MaxBandwidth   uint
	MaxClock       uint
	MaxCore        uint
	MaxFreeMemory  uint64
	MaxPower       uint
	MaxTotalMemory uint64
}

func (s *Data) Clone() framework.StateData {   // StateData is a generic type for arbitrary data stored in CycleState. similar to empty interface
	c := &Data{
		Value: s.Value,
	}
	return c
}


/*
遍历集群中所有可用节点的所有可用card, 在满足number, memory, clock等标签参数要求前提下，求各参数（如clock, bandwith)的最大值
 */
func CollectMaxValues(state *framework.CycleState, pod *v1.Pod, scvList scv.ScvList) *framework.Status {
	//CycleState provides a mechanism for plugins to store and retrieve arbitrary data
	//Status indicates the result of running a plugin
	data := Data{
		Value: MaxValue{
			MaxBandwidth:   1,
			MaxClock:       1,
			MaxCore:        1,
			MaxFreeMemory:  1,
			MaxPower:       1,
			MaxTotalMemory: 1,
		},
	}

	for _, item := range scvList.Items {   // 一个scv对应一个节点，因此scvList表示集群中所有节点
		// APIResource struct have the DeepCopy() function, refer to https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#APIResource
		s := item.DeepCopy()
		if ok, number := filter.PodFitsNumber(pod, s); ok {
			isFitsMemory, memory := filter.PodFitsMemory(number, pod, s)
			isFitsClock, clock := filter.PodFitsClock(number, pod, s)
			if isFitsMemory && isFitsClock {  // 节点满足要求
				for _, card := range s.Status.CardList {
					if card.FreeMemory >= memory && card.Clock >= clock {   // 节点的GPU卡满足要求
						ProcessMaxValueWithCard(card, &data)
					}
				}
			}
		}

	}
	state.Lock()
	state.Write("Max", &data)   // This function is not thread safe. In multi-threaded code, lock should be acquired first.
	defer state.Unlock()
	return framework.NewStatus(framework.Success, "")
}

func ProcessMaxValueWithCard(card scv.Card, data *Data) {
	if card.FreeMemory > data.Value.MaxFreeMemory {
		data.Value.MaxFreeMemory = card.FreeMemory
	}
	if card.Clock > data.Value.MaxClock {
		data.Value.MaxClock = card.Clock
	}
	if card.TotalMemory > data.Value.MaxTotalMemory {
		data.Value.MaxTotalMemory = card.TotalMemory
	}
	if card.Bandwidth > data.Value.MaxBandwidth {
		data.Value.MaxBandwidth = card.Bandwidth
	}
	if card.Core > data.Value.MaxCore {
		data.Value.MaxCore = card.Core
	}
	if card.Power > data.Value.MaxPower {
		data.Value.MaxPower = card.Power
	}
}

