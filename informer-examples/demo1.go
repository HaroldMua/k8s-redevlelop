package main

import (
	"flag"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"log"
	"path/filepath"
	"time"
)

func main() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		/*
			https://pkg.go.dev/flag#String
			func (f *FlagSet) String(name string, value string, usage string) *string
			String defines a string flag with specified name, default value, and usage string. The return value is the address of a string variable that stores the value of the flag.
		*/
		kubeconfig = flag.String("kubeconfig", filepath.Join(".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	// 创建stopCh对象，该对象用于在程序进程退出之前通知informer提前退出，因为informer是一个持久运行的goroutine
	stopCh := make(chan struct{})
	defer close(stopCh)
	//表示每分钟进行一次resync，resync会周期性地执行List操作
	sharedInformers := informers.NewSharedInformerFactory(clientset, time.Minute)

	informer := sharedInformers.Core().V1().Pods().Informer()

	//informer.AddEventHandler函数可以为Pod资源添加资源事件回调方法
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			mObj := obj.(v1.Object)   // obj.()?
			log.Printf("New pod added to store: %s", mObj.GetName())
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oObj := oldObj.(v1.Object)
			nObj := newObj.(v1.Object)
			log.Printf("%s pod updated to %s", oObj.GetName(), nObj.GetName())
		},
		DeleteFunc: func(obj interface{}) {
			mObj := obj.(v1.Object)
			log.Printf("pod deleted from store: %s", mObj.GetName())
		},
	})

	informer.Run(stopCh)
}
