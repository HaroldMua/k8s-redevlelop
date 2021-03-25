# CGO_ENABLED=0表示go采用纯静态编译
# GOOS=linux表示编译Linux可执行程序
build:
	CGO_ENABLED=0 GOOS=linux go build -o website-controller -a pkg/website-controller.go

image: build
	docker build -t haroldmua/website-controller .

push: image
	docker push haroldmua/website-controller