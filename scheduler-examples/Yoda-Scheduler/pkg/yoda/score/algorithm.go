package score

import (
	"errors"
	scv "SCV/api/v1"
	"Yoda-Scheduler/pkg/yoda/collection"
	"Yoda-Scheduler/pkg/yoda/filter"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

// 设置权重值
const (
	BandwidthWeight   = 1
	ClockWeight       = 1
	CoreWeight        = 1
	PowerWeight       = 1
	FreeMemoryWeight  = 2
	TotalMemoryWeight = 1
	ActualWeight      = 2

	AllocateWeight = 3
)

func CalculateScore (s *scv.Scv, state *framework.CycleState, pod *v1.Pod, info *framework.NodeInfo) (uint64, error) {
	// 若没有参考，自己是不可能想到使用state.Read()的，
	d, err := state.Read("Max")   // refer to: https://pkg.go.dev/k8s.io/kubernetes/pkg/scheduler/framework#StateData
	if err != nil {
		klog.V(3).Infof("Error Get CycleState Info: %v", err)
		return 0, err
	}
	data, ok := d.(*collection.Data)   // d是interface类型，但是这个写法是几个意思？
	if !ok {
		return 0, errors.New("The Type is not Data ")
	}
	return CalculateBasicScore(data.Value, s, pod) + CalculateAllocateScore(info, s) + CalculateActualScore(s), nil
}

/*
计算节点所有卡的分数
例如，要求如下：
labels:
	scv/number: "2"
	scv/memory: "8000"
	scv/clock: "5705"

 */
func CalculateBasicScore(value collection.MaxValue, scv *scv.Scv, pod *v1.Pod) uint64 {
	var cardScore uint64
	if ok, number := filter.PodFitsNumber(pod, scv); ok {   // 返回scv/number
		isFitsMemory, memory := filter.PodFitsMemory(number, pod, scv)  // whether there are at least "number" card fits memory, reture scv/memory
		isFitsClock, clock := filter.PodFitsClock(number, pod, scv)   // whether there are at least "number" card fits clock, reture scv/clock
		if isFitsMemory && isFitsClock {  // 节点上至少有“number”个card同时满足memory和clock
			for _, card := range scv.Status.CardList {
				if card.FreeMemory >= memory && card.Clock >= clock {   // 节点上其他的卡是否满足memory和clock要求，例如，要求2个卡，节点有3个卡
					cardScore += CalculateCardScore(value, card)
				}
			}
		}
	}
	return cardScore
}

// 计算节点单个卡的分数
func CalculateCardScore(value collection.MaxValue, card scv.Card) uint64 {
	var (
		bandwith    = card.Bandwidth * 100 / value.MaxBandwidth
		clock       = card.Clock * 100 / value.MaxBandwidth
		core        = card.Core * 100 / value.MaxCore
		power       = card.Power * 100 / value.MaxPower
		freeMemory  = card.FreeMemory * 100 / value.MaxFreeMemory
		totalMemory = card.TotalMemory * 100 / value.MaxTotalMemory
	)
	return uint64(bandwith * BandwidthWeight + clock * ClockWeight + core * CoreWeight + power * PowerWeight) +
		freeMemory * FreeMemoryWeight + totalMemory * TotalMemoryWeight
}

// 计算节点所有卡的内存实际余量之和分数（计算实际余量的意义是，该卡的内存资源可能被非集群的pod任务占用，如与k8s集群无关的用户程序）
// 但是，关键是看单卡的余量是否满足memory要求，因此该分数仅作参考（单卡节点时，有效）
func CalculateActualScore(scv *scv.Scv) uint64 {
	return (scv.Status.FreeMemorySum * 100 / scv.Status.TotalMemorySum) * ActualWeight
}

// 计算节点所有卡的调度层面的内存余量之和分数
func CalculateAllocateScore(info *framework.NodeInfo, scv *scv.Scv) uint64 {
	allocateMemorySum := uint64(0)
	for _, pod := range info.Pods {
		if mem, ok := pod.Pod.GetLabels()["scv/memory"]; ok {
			allocateMemorySum += filter.StrToUint64(mem)
		}
	}
	if scv.Status.TotalMemorySum < allocateMemorySum {
		return 0
	}
	return ((scv.Status.TotalMemorySum - allocateMemorySum) * 100 / scv.Status.TotalMemorySum) * AllocateWeight
}