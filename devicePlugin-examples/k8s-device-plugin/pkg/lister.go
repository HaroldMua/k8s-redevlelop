package pkg

import (
	"github.com/kubevirt/device-plugin-manager/pkg/dpm"
)

type Lister struct{}

func (Lister) GetResourceNamespace() string {
	return "extend-k8s.io"
}

func (Lister) Discover(pluginListCh chan dpm.PluginNameList) {
	pluginListCh <- dpm.PluginNameList{"example"}
}

func (Lister) NewPlugin(deviceID string) dpm.PluginInterface {
	return &ExamplePlugin{}
}