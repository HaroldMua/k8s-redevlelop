package main

import (
	"fmt"
	"net/http"
	"encoding/json"
	"io"
	"log"
	"io/ioutil"
	"strings"
	"k8s-website-controller/pkg/v1"
)

/*
通过HTTP GET请求打开的连接，API服务器将针对任何websites对象的每个更改发送监听事件(watch event)
每次创建新的websites对象时，API服务器都会发送ADDED监听事件。当控制器收到这样的事件时，就会在该监听事件所包含的websites对象
中提取网站名称和Git存储库的URL，然后将它们的JSON清单发布到API服务器，来创建Deployment和Service对象
 */
func main() {
	log.Println("website-controller started.")
	for {
		resp, err := http.Get("http://localhost:8001/apis/extensions.example.com/v1/websites?watch=true")
		if err != nil {
			panic(err)
		}
		//使用defer来关闭，这会在封闭函数（main）结束时执行
		defer resp.Body.Close()


		decoder := json.NewDecoder(resp.Body)
		for {
			var event v1.WebsiteWatchEvent
			if err := decoder.Decode(&event); err == io.EOF {
				break
			} else if err != nil{
				log.Fatal(err)
			}

			log.Printf("Received watch event: %s: %s: %s\n", event.Type, event.Object.Metadata.Name, event.Object.Spec.GitRepo)

			if event.Type == "ADDED" {
				createWebsite(event.Object)
			} else if event.Type == "DELETED" {
				deleteWebsite(event.Object)
			}
		}
	}
}

func createWebsite(website v1.Website) {
	createResource(website, "api/v1", "services", "service-template.json")
	createResource(website, "apis/apps/v1", "deployments", "deployment-template.json")
}

func deleteWebsite(website v1.Website) {
	deleteResource(website, "api/v1", "services", getName(website))
	deleteResource(website, "apis/apps/v1", "deployments", getName(website))
}

func createResource(webserver v1.Website, apiGroup string, kind string, filename string) {
	log.Printf("Creating %s with name %s in namespace %s", kind, getName(webserver), webserver.Metadata.Namespace)
	templateBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	/*
	func Replace(s, old, new string, n int) string.
	If n < 0, there is no limit on the number of replacements.
	 */
	template := strings.Replace(string(templateBytes), "[NAME]", getName(webserver), -1)
	template = strings.Replace(template, "[GIT-REPO]", webserver.Spec.GitRepo, -1)

	//客户端通过创建到API服务器（启动kubectl proxy命令，可通过localhost:8001访问API服务器）的HTTP连接来Create资源
	resp, err := http.Post(fmt.Sprintf("http://localhost:8001/%s/namespaces/%s/%s/", apiGroup, webserver.Metadata.Namespace, kind), "application/json", strings.NewReader(template))
	if err != nil {
		log.Fatal(err)
	}
	log.Println("response Status:", resp.Status)
}

func deleteResource(webserver v1.Website, apiGroup string, kind string, name string) {
	log.Printf("Deleting %s with name %s in namespace %s", kind, getName(webserver), webserver.Metadata.Namespace)
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("http://localhost:8001/%s/namespaces/%s/%s/%s", apiGroup, webserver.Metadata.Namespace, kind, name), nil)
	if err != nil {
		log.Fatal(err)
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Println("response Status:", resp.Status)
}

func getName(websit v1.Website) string {
	return websit.Metadata.Name + "-website"
}