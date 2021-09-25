package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	//podresourcesapi "k8s.io/kubelet/pkg/apis/podresources/v1"
	"k8s.io/kubernetes/pkg/kubelet/apis/podresources/v1alpha1"
)


type PodInfo struct {
	PodName   string
	Namespace string
}

type PodCollector struct {
	PodList []*PodInfo
}

func main() {
	socketPath := "/var/lib/kubelet/pod-resources/kubelet.sock"
	conn, cleanup, err := connectToServer(socketPath)
	if err != nil {
		log.Printf("Can not connect to: %v", socketPath)
	}
	defer cleanup()

	fmt.Println("successfully connect to socketPath")
	listPodResp, err := ListPods(conn)
	if err != nil {
		log.Printf("Can not list pod resp, :%v", err)
	}

	podCollector := &PodCollector{}
	for _, pod := range listPodResp.GetPodResources() {
		podInfo := &PodInfo{}
		podInfo.PodName = pod.Name
		podInfo.Namespace = pod.Namespace

		podCollector.PodList = append(availablePodInfo, podInfo)
	}
}

func connectToServer(socket string) (*grpc.ClientConn, func(), error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, socket, grpc.WithInsecure(), grpc.WithBlock(),
		grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
			return net.DialTimeout("unix", addr, timeout)
		}),
	)

	if err != nil {
		return nil, func() {}, fmt.Errorf("failure connecting to %s: %v", socket, err)
	}

	return conn, func() { conn.Close() }, nil
}

func ListPods(conn *grpc.ClientConn) (*podresourcesapi.ListPodResourcesResponse, error) {
	client := podresourcesapi.NewPodResourcesListerClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
	defer cancel()

	resp, err := client.List(ctx, &podresourcesapi.ListPodResourcesRequest{})
	if err != nil {
		return nil, fmt.Errorf("failure getting pod resources %v", err)
	}

	return resp, nil
}