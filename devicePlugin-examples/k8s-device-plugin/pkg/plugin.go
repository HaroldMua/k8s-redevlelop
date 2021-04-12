package pkg

import (
	"context"
	"math/rand"
	"time"

	"github.com/thanhpk/randstr"
	. "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

type ExamplePlugin struct {}

func (dp *ExamplePlugin) ListAndWatch(e *Empty, s DevicePlugin_ListAndWatchServer) error {
	s.Send(&ListAndWatchResponse{Devices: randomDevices()})
	for {
		time.Sleep(5 * time.Second)
		s.Send(&ListAndWatchResponse{Devices: randomDevices()})
	}
}

func (dp *ExamplePlugin) Allocate(c context.Context, r *AllocateRequest) (*AllocateResponse, error) {
	envs := map[string]string{"K8S_DEVICE_PLUGIN_EXAMPLE": randstr.Hex(16)}
	responses := []*ContainerAllocateResponse{{Envs: envs}}

	return &AllocateResponse{ContainerResponses: responses}, nil
}

func (ExamplePlugin) GetDevicePluginOptions(context.Context, *Empty) (*DevicePluginOptions, error) {
	return nil, nil
}

func (ExamplePlugin) PreStartContainer(context.Context, *PreStartContainerRequest) (*PreStartContainerResponse, error) {
	return nil, nil
}

func (dp *ExamplePlugin) GetPreferredAllocation(context.Context, *PreferredAllocationRequest) (*PreferredAllocationResponse, error) {
	return nil, nil
}

func randomDevices() []*Device {
	devices := make([]*Device, 0)
	for i := 0; i < rand.Intn(5) + 1; i++ {
		devices = append(devices, &Device{
			ID:     randstr.Hex(16),
			Health: Healthy,
		})
	}
	return devices
}
