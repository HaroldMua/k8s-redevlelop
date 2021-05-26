package main

import (
	"Yoda-Scheduler/pkg/register"
	"fmt"
	"k8s.io/component-base/logs"
	"math/rand"
	"os"
	"time"
)

func main() {
	// 确保每次取的随机数都是随机的
	rand.Seed(time.Now().UnixNano())
	command := register.Register()

	logs.InitLogs()
	defer logs.FlushLogs()

	if err := command.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

}