package client

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

/*
ClientSet仅能访问k8s自身内置的资源（即客户端集合内的资源），
不能直接访问CRD自定义资源。如果需要ClientSet访问CRD自定义资源，
可通过client-gen代码生成器重新生成ClientSet,在ClientSet集合中自动生成与CRD操作相关的接口
 */
func Connect() (*kubernetes.Clientset, *rest.Config) {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return clientset, config
}