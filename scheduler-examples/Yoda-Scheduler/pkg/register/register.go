package register

import (
	"github.com/spf13/cobra"
	"k8s.io/kubernetes/cmd/kube-scheduler/app"
	"Yoda-Scheduler/pkg/yoda"
)

func Register() *cobra.Command {
	return app.NewSchedulerCommand(
		app.WithPlugin(yoda.Name, yoda.New),
	)
}