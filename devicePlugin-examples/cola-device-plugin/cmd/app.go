package main

import (
	"os"
	"path/filepath"
	"time"

	"cola-device-plugin/pkg/server"
	log "github.com/sirupsen/logrus"
	//"gopkg.in/fsnotify.v1"
	"github.com/fsnotify/fsnotify"
)

func main() {
	log.Info("cola device plugin starting...")
	colaSrv := server.NewColaServer()
	go colaSrv.Run()

	// 向kubelet注册
	if err := colaSrv.RegisterToKubelet(); err != nil {
		log.Fatalf("register to kubelet error: %v", err)
	} else {
		log.InfoLn("register to kubelet successfully")
	}

    /*
    使用 fsnotify 类似的库监控 kubelet.sock 的重新创建事件。如果重新创建了，则认为 kubelet 是重启了，
    我们需要重新向 kubelet 注册 device plugin。
     */
    devicePluginSocket := filepath.Join(server.DevicePluginPath, server.KubeletSocket)
    log.Info("device plugin socket name:", devicePluginSocket)
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
    	log.Error("Failed to cerate FS watcher")
    	os.Exit(1)
    }

    defer watcher.Close()
    err = watch.Add(server.DevicePluginPath)
    if err != nil {
    	log.Error("watch kubelet error")
    	return
	}

	log.Info("watching kubelet.sockt")
    for {
    	select {
    	case event := <-watcher.Events:
			log.Infof("watch kubelet events: %s, event name: %s, isCreate: %v", event.Op.String(), event.Name, event.Op&fsnotify.Create == fsnotify.Create)
			if evnet.Name == devicePluginSocket && event.Op&fsnotify.Create == fsnotify.Create {
				time.Sleep(time.Second)
				log.Fatalf("inotify: %s created, restarting.", devicePluginSocket)
			}
		case err := <-watcher.Errors:
			log.Fatalf("inotify: %s", err)
		}
	}
}