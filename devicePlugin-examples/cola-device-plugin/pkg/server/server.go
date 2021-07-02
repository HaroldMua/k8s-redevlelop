package server

import (
	"context"
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	pluginapi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1beta1"
)

const (
	resourceName        string = "myway5.com/cola"
	defaultColaLocation string = "/etc/colas"
	colaSocket          string = "cola.sock"
	// KubeletSocket kubelet 监听 unix 的名称
	KubeletSocket string = "kubelet.sock"
	// DevicePluginPath 默认位置
	DevicePluginPath string = "/var/lib/kubelet/device-plugins/"
)

// ColaServer 是一个 device plugin server
type ColaServer struct {
	srv         *grpc.Server
	devices     map[string]*pluginapi.Device
	notify      chan bool
	ctx         context.Context
	cancel      context.CancelFunc
	restartFlag bool // 本次是否是重启
}

// NewColaServer 实例化 colaServer，在cmd/server/app.go中调用实例化
func NewColaServer() *ColaServer {
	ctx, cancel := context.WithCancel(context.Background())
	return &ColaServer{
		srv: grpc.NewServer(grpc.EmptyServerOptions{}),
		devices: make(map[string]*pluginapi.Device),
		notify: make(chan bool),
		ctx: ctx,
		cancel: cancel,
		restartFlag: false,
	}
}

// Run 运行服务
func (s *ColaServer) run() error {
	// 发现本地设备
	err := s.listDevice()
	if err != nil {
		log.Fatalf("list device error: %v", err)
	}

	go func() {
		err := s.watchDevice()
		if err != nil {
			log.Println("watch device error")
		}
	}()

	/*
	Refer to: https://pkg.go.dev/k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1beta1#RegisterDevicePluginServer
	func RegisterDevicePluginServer(s *grpc.Server, srv DevicePluginServer)

	Refer to: https://pkg.go.dev/k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1beta1#DevicePluginServer
	type DevicePluginServer interface {
		// GetDevicePluginOptions returns options to be communicated with Device
		// Manager
		GetDevicePluginOptions(context.Context, *Empty) (*DevicePluginOptions, error)
		// ListAndWatch returns a stream of List of Devices
		// Whenever a Device state change or a Device disappears, ListAndWatch
		// returns the new list
		ListAndWatch(*Empty, DevicePlugin_ListAndWatchServer) error
		// Allocate is called during container creation so that the Device
		// Plugin can run device specific operations and instruct Kubelet
		// of the steps to make the Device available in the container
		Allocate(context.Context, *AllocateRequest) (*AllocateResponse, error)
		// PreStartContainer is called, if indicated by Device Plugin during registeration phase,
		// before each container start. Device plugin can run device specific operations
		// such as reseting the device before making devices available to the container
		PreStartContainer(context.Context, *PreStartContainerRequest) (*PreStartContainerResponse, error)
	}
	 */

	//DevicePluginServer是一个接口类型，在这里，由于结构体ColaServer实现了接口DevicePluginServer的所有方法，且s是实例化的ColaServer，
	// 因此，相当于把ColaServer实例赋值给DevicePluginServer
	pluginapi.RegisterDevicePluginServer(s.srv, s)
	err = syscall.Unlink(DevicePluginPath + colaSocket)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	l, err := net.Listen("unix", DevicePluginPath + colaSocket)
	if err != nil {
		return err
	}

	go func() {
		lastCrashTime := time.Now()
		restartCount := 0
		for {
			log.Printf("start GRPC server for '%s'", resourceName)
			err = s.srv.Serve(l)
			if err != nil {
				break
			}

			log.Printf("GRPC server for '%s' crashed wiht error: %v", resourceName, err)

			if restartCount > 5 {
				log.Fatal("GRPC server for '%s' has repeatedly crashed recently. Quitting", resourceName)
			}
			timeSinceLastCrash = time.Since(lastCrashTime).Seconds()
			lastCrashTime = time.Now()
			if timeSinceLastCrash > 3600 {
				restartCount = 1
			} else {
				restartCount++
			}
		}
	}()

	// Wait for server to start by lauching a blocking connection
	conn, err := s.dial(colaSocket, 5*time.Second)
	if err != nil {
		return err
	}
	conn.Close()

	return nil
}

// RegisterToKubelet 向kubelet注册device plugin
func (s *ColaServer) RegisterToKubelet() error {
	socketFile := filepath.Join(DevicePluginPath + KubeletSocket)

	conn, err := s.dial(socketFile, 5*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pluginapi.NewRegistrationClient(conn)
	req := &pluginapi.RegisterRequest{
		Version:      pluginapi.Version,
		Endpoint:     path.Base(DevicePluginPath + colaSocket),
		ResourceName: resourceName,
	}
	log.Infof("Register to kubelet with endpoint %s", req.Endpoint)
	_, err = client.Register(context.Background(), req)
	if err != nil {
		return err
	}

	return nil
}

// GetDevicePluginOptions returns options to be communicated with Device
// Manager
func (s *ColaServer) GetDevicePluginOptions(ctx context.Context, e *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	log.Infoln("GetDevicePluginOptions called")
	return &pluginapi.DevicePluginOptions{PreStartRequired: true}, nil
}

// ListAndWatch returns a stream of List of Devices
// Whenever a Device state change or a Device disappears, ListAndWatch
// returns the new list
func (s *ColaServer) ListAndWatch(e *pluginapi.Empty, srv pluginapi.DevicePlugin_ListAndWatchServer) error {
	log.Infoln("ListAndWatch called")
	devs := make([]*pluginapi.Device, len(s.devices))

	i := 0
	for _, dev := range s.devices {
		devs[i] = dev
		i++
	}

	err := srv.Send(&pluginapi.ListAndWatchResponse{Devices: devs})
	if err != nil {
		log.Errorf("ListAndWatch send device error: %v", err)
		return err
	}

	// 更新device list
	for {
		log.Infoln("waiting for device change")
		select {
		case <-s.notify:
			log.Infoln("开始更新device list, 设备数：", len(s.devices))
			devs := make([]*pluginapi.Device, len(s.devices))

			i := 0
			for _, dev := range s.devices {
				devs[i] = dev
				i++
			}

			srv.Send(&pluginapi.ListAndWatchResponse{Devices: devs})
		case <-s.ctx.Done():
			log.Info("ListAndWatch exit")
			return nil
		}

	}

}

// Allocate is called during container creation so that the Device
// Plugin can run device specific operations and instruct Kubelet
// of the steps to make the Device available in the container
func (s *ColaServer) Allocate(ctx context.Context, reqs *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	log.Infoln("Allocate called")
	resps := &pluginapi.AllocateResponse{}
	for _, req := range reqs.ContainerRequests {
		log.Infof("received request: %v", strings.Join(req.DevicesIDs, ","))
		resp := pluginapi.ContainerAllocateResponse{
			Envs: map[string]string{
				"COLA_DEVICES": strings.Join(req.DevicesIDs, ","),
			},
		}
		resps.ContainerResponses = append(resps.ContainerResponses, &resp)
	}

	return resps, nil
}

// PreStartContainer is called, if indicated by Device Plugin during registeration phase,
// before each container start. Device plugin can run device specific operations
// such as reseting the device before making devices available to the container
func (s *ColaServer) PreStartContainer(ctx context.Context, req *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	log.Infoln("PreStartContainer called")
	return &pluginapi.PreStartContainerResponse{}, nil
}

// listDevice 从节点上发现设备
//定义的 myway5.com/cola 资源用 /etc/colas 下的文件代表
func (s *ColaServer) listDevice() error {
	dir, err := ioutil.ReadDir(defaultColaLocation)
	if err != nil {
		return err
	}

	for _, f := range dir {
		if f.IsDir() {
			continue
		}

		// Package md5 implements the MD5 hash algorithm as defined in RFC 1321.
		sum := md5.Sum([]byte(f.Name()))    // 要修改字符串，需要先将f.Name()转换成[]rune或[]byte，完成后再转换为string。
		s.devices[f.Name()] = &pluginapi.Device{
			ID:     string(sum[:]),
			Health: pluginapi.Healthy,
		}
		log.Infof("find device: '%s'", f.Name())
	}

	return nil
}

func (s *ColaServer) watchDevice() error {
	log.Infof("watch device")
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("NewWatcher error: %v", err)
	}
	defer w.Close()

	done := make(chan bool)
	go func() {
		defer func() {
			done <- true   // 当Goroutine执行完，向通道done发送true
			log.Infof("watch device exist")
		}()
		for {
			select {
			case event, ok := <-w.Events:   // 定义evert为单向通道
			    if !ok {
			    	continue
				}
				log.Infoln("device event:", event.Op.String())

			    if event.Op&fsnotify.Create == fsnotify.Create {
					// 创建文件，增加 device
			    	sum := md5.Sum([]byte(event.Name))
			    	s.devices[event.Name] = &pluginapi.Device{
			    		ID:     string(sum[:]),
			    		Health: pluginapi.Healthy,
					}
					s.notify <- true
					log.Infoln("new device find:", event.Name)
				} else if event.Op&fsnotify.Remove == fsnotify.Remove {
					// 删除文件， 删除device
					delete(s.devices, event.Name)   // Go语言的 delete(map, key) 函数用于删除集合的某个元素，参数为 map 和其对应的 key。
					s.notify <- true
					log.Infoln("device deleted:", event.Name)
				}

			case err, ok := <-w.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)

			case <-s.ctx.Done():
				break
			}
		}
	}()

	<-done   // // 接受另一个Goroutine的值, 标志运行结束

	return nil
}

func (s *ColaServer) dial(unixSocketPath string, timeout time.Duration) (*grpc.ClientConn, error) {
	c, err := grpc.Dial(unixSocketPath,
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithTimeout(timeout),
		grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
			return net.DialTimeout("unix", addr, timeout)
		}),
	)

	if err != nil {
		return nil, err
	}

	return c, nil
}