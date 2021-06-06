package collector

import (
	"context"
	"os"
	"reflect"
	"sort"
	"time"

	"github.com/NVIDIA/gpu-monitoring-tools/bindings/go/nvml"

	apierror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "SCV/api/v1"
	"SCV/pkg/log"
)

type Collector struct {
	// CRD info
	cache  cache.Cache
	client client.Client

	nodeName string

	// GPU info
	cardList       v1.CardList
	cardNumber     uint
	FreeMemorySum  uint64
	TotalMemorySum uint64

	updateInterval int64
}

func (c *Collector) CountGPU() {
	err := nvml.Init()
	if err != nil {
		log.ErrPrint(err)
	}
	defer func() {
		if err := nvml.Shutdown(); err != nil {
			log.ErrPrint(err)
		}
	}()

	count, err := nvml.GetDeviceCount()
	if err != nil {
		log.ErrPrint(err)
	}
	c.cardNumber = count
}

func (c *Collector) UpdateGPU() {
	newCardList := make(v1.CardList, 0)
	err := nvml.Init()
	if err != nil {
		log.ErrPrint(err)
	}
	defer func() {
		if err := nvml.Shutdown(); err != nil {
			log.ErrPrint(err)
		}
	}()

	c.CountGPU()

	for i := uint(0); i < c.cardNumber; i++ {
		device, err := nvml.NewDevice(i)
		if err != nil {
			log.ErrPrint(err)
		}
		health := "Healthy"
		status, err := device.Status()
		if err != nil {
			log.ErrPrint(err)
			health = "Unhealthy"   // health变量在上面已经定义了类型
		}
		newCardList = append(newCardList, v1.Card{
			ID:          i,
			Health:      health,
			Model:       *device.Model,
			Power:       *device.Power,
			TotalMemory: *device.Memory,
			Clock:       *device.Clocks.Memory,
			FreeMemory:  *status.Memory.Global.Free,
			Core:        *device.Clocks.Cores,
			Bandwidth:   *device.PCI.Bandwidth,
		})
	}

	sort.Sort(newCardList)
	if len(c.cardList) == 0 || reflect.DeepEqual(c.cardList, newCardList) {   // func DeepEqual(x, y interface{}) bool, DeepEqual reports whether x and y are “deeply equal,”
		c.cardList = newCardList
	}

	total, free := uint64(0), uint64(0)
	for _, card := range newCardList {
		total += card.TotalMemory
		free += card.FreeMemory
	}
	c.TotalMemorySum = total
	c.FreeMemorySum = free
	c.cardList = newCardList
}

func (c *Collector) createScv() error {
	scv := v1.Scv{
		ObjectMeta: metav1.ObjectMeta{
			Name: c.nodeName,
		},
		Spec: v1.ScvSpec{
			UpdateInterval: c.updateInterval,
		},
	}
	err := c.client.Create(context.Background(), &scv)   // https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/client#Client
	if err != nil && !apierror.IsAlreadyExists(err) {
		return err
	}

	return nil
}

func (c *Collector) NeedUpdate(status v1.ScvStatus) bool {
	if status.UpdateTime == nil {
		log.Print("CardList is Null, needs update.")
		return true
	}

	if status.TotalMemorySum != c.TotalMemorySum {
		log.Print("Total memory changed, needs update.")
		return true
	}

	if status.FreeMemorySum != c.FreeMemorySum {
		log.Print("Free memory changed, needs update.")
		return true
	}
	if status.CardNumber != c.cardNumber {
		log.Print("Card Number changed, needs update.")
		return true
	}
	if !reflect.DeepEqual(status.CardList, c.cardList) {
		log.Print("Card List changed, needs update.")
		return true
	}
	return false
}

func (c *Collector) Process() {
	interval := time.Duration(c.updateInterval) * time.Millisecond
	ticker := time.NewTicker(interval)
	for {
		<-ticker.C
		// update the info of GPU
		c.UpdateGPU()

		currentScv := v1.Scv{}

		/*
		original sytle:

		key := types.NamespacedName{
			Name: c.nodeName,
		}
		err := c.client.Get(context.TODO(), key, &currentScv)
		*/

		err := c.client.Get(context.Background(), client.ObjectKey{
			Name: c.nodeName,
		}, &currentScv)
		if err != nil {
			log.ErrPrint(err)
			continue
		}

		// update the status of Scv, if there are no changes in GPU info, don't need update status
		if c.NeedUpdate(currentScv.Status) {
			updateScv := currentScv.DeepCopy()  // APIResource struct have the DeepCopy() function, refer to https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#APIResource
			updateScv.Status = v1.ScvStatus{
				CardList:       c.cardList,
				TotalMemorySum: c.TotalMemorySum,
				FreeMemorySum:  c.FreeMemorySum,
				CardNumber:     c.cardNumber,
				UpdateTime:     &metav1.Time{
					Time: time.Now(),
				},
			}

			if err := c.client.Update(context.Background(), updateScv); err !=nil {
				log.ErrPrint(err)
			}
		}
	}

}

func NewCollector(interval int64, client client.Client, cache cache.Cache) *Collector {
	return &Collector{
		cardList:       make(v1.CardList, 0),
		cardNumber:     0,
		updateInterval: interval,
		client:         client,
		cache:          cache,
	}
}

func StartCollector(c *Collector) {
	// Init CRD & Set config
	c.nodeName = os.Getenv("NODE_NAME")
	if err := c.createScv(); err != nil {
		panic(err)
	}
	c.Process()
}

/*
Logic here is:

NewCollector --> StartCollector
	createScv
	Process
		UpdateGPU
			CountGPU
		NeedUpdate

 */