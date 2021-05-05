module operator-demo-with-code-generator

go 1.15

require (
	github.com/golang/glog v0.0.0-20210429001901-424d2337a529
	k8s.io/api v0.21.0
	k8s.io/apimachinery v0.21.0
	k8s.io/code-generator v0.21.0
	k8s.io/client-go v0.21.0 // 手动指定版本，不然报错
)

