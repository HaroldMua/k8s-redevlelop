# SCV

SCV is a distributed cluster GPU sniffer, partly generated with the help of [kubebuilder](https://book.kubebuilder.io/). Refered to [this](https://github.com/NJUPT-ISL/SCV). 

Three important files is:
- main.go
- api/v1/scv_types.go
- pkg/collector/collector.go

Logic here is:

```bazaar
NewCollector --> StartCollector
  createScv
  Process
    UpdateGPU
	  CountGPU
    NeedUpdate
```

## 自定义Scv资源

Scv yaml文件的metadata, spec, status只需自定义后两个，metadata在使用kuberbuilder创建资源时已经定义好。

spec属性：
- updateInterval

status属性：
- cardList:
  - bandwith
  - clock
  - core
  - freeMemory
  - health
  - id
  - model
  - power
  - totalMemory
- cardNumber
- freeMemorySum
- totalMemorySum
- updateTime

cardList使用slice表示，card使用结构体表示

## GPU metrics that SCV can monitor
- Core Frequency
- Model
- Free Memory 
- Total Memory 
- Memory Frequency
- Bandwidth
- Power
- GPU Number

## CRD Example
```yaml
apiVersion: core.run-linux.com/v1
kind: Scv
metadata:
  creationTimestamp: "2021-05-21T01:55:12Z"
  generation: 2
  name: gpu02-poweredge-t420
  resourceVersion: "22929326"
  selfLink: /apis/core.run-linux.com/v1/scvs/gpu02-poweredge-t420
  uid: bd0121d9-46e4-4c52-8cb7-caed885a9580
spec:
  updateInterval: 1000
status:
  cardList:
  - bandwidth: 8000
    clock: 5705
    core: 1911
    freeMemory: 12196
    health: Healthy
    id: 0
    model: TITAN Xp
    power: 250
    totalMemory: 12196
  cardNumber: 1
  freeMemorySum: 12196
  totalMemorySum: 12196
  updateTime: "2021-05-21T01:55:13Z"
```

### Get Started
- Ensure that the nvidia container runtime and the nvidia driver are installed on each kubernetes worker node. See [nvidia-docker](https://github.com/NVIDIA/nvidia-docker#quickstart)
for more details.
    -  Ubuntu 
    
       ```shell
       # Add the package repositories
       $ distribution=$(. /etc/os-release;echo $ID$VERSION_ID)
       $ curl -s -L https://nvidia.github.io/nvidia-docker/gpgkey | sudo apt-key add -
       $ curl -s -L https://nvidia.github.io/nvidia-docker/$distribution/nvidia-docker.list | sudo tee /etc/apt/sources.list.d/nvidia-docker.list
            
       $ sudo apt-get update && sudo apt-get install -y nvidia-container-toolkit nvidia-container-runtime
       $ sudo systemctl restart docker
        ```
    - Centos
    
        ```shell
        $ distribution=$(. /etc/os-release;echo $ID$VERSION_ID)
        $ curl -s -L https://nvidia.github.io/nvidia-docker/$distribution/nvidia-docker.repo | sudo tee /etc/yum.repos.d/nvidia-docker.repo
            
        $ sudo yum install -y nvidia-container-toolkit nvidia-container-runtime
        $ sudo systemctl restart docker
        ```
- Enable the nvidia-container-runtime as docker default runtime on each kubernetes worker node.

    You need to modify `/etc/docker/daemon.json` to the following content on each worker node：
    ```json
        {
            "default-runtime": "nvidia",
            "runtimes": {
                "nvidia": {
                    "path": "/usr/bin/nvidia-container-runtime",
                    "runtimeArgs": []
                }
            },
            "exec-opts": ["native.cgroupdriver=systemd"],
            "log-driver": "json-file",
            "log-opts": {
              "max-size": "100m"
            },
            "storage-driver": "overlay2",
            "registry-mirrors": ["https://registry.docker-cn.com"]
        }
    ```
- First, install CRD on kubernetes cluster by:
    ```
    make install
    ```
  
- Now you can check if the CRD was successfully installed by:
    ```
    kubectl get CRD 
    ```
  
- Deploy the SCV into your kubernetes cluster:
    ```shell
    kubectl apply -f  https://raw.githubusercontent.com/NJUPT-ISL/SCV/master/deploy/deploy.yaml
    ```

- Undeploy the SCV and uninstall CRD
    ```shell
    kubectl apply -f  https://raw.githubusercontent.com/NJUPT-ISL/SCV/master/deploy/deploy.yaml
    make uninstall
    ```  
- After deploying the daemonset, the SCV resources will be created according to the corresponding node name.

## Next step
- If the SCV resource is deleted by hand, the daemonset controller should recreate the SCV automatically.
