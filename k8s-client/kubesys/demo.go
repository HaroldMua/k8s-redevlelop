package main

import (
	"encoding/json"
	"fmt"
	"github.com/kubesys/kubernetes-client-go/pkg/kubesys"
)

const(
	DefaultMasterUrl = "https://192.168.1.24:6443"
	DefaultToken     = "eyJhbGciOiJSUzI1NiIsImtpZCI6ImF5R1h4SFJkR0x1UzNMVklkd01tZnRxdUxTYkdCc2dwdnlXLUFoV1ZUM3MifQ.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJrdWJlLXN5c3RlbSIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VjcmV0Lm5hbWUiOiJrdWJlcm5ldGVzLWNsaWVudC10b2tlbi1rbndjZiIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VydmljZS1hY2NvdW50Lm5hbWUiOiJrdWJlcm5ldGVzLWNsaWVudCIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VydmljZS1hY2NvdW50LnVpZCI6ImJlMDA0NTFiLThmNDUtNDU1NC1hNWQ4LTFmODRhNzA4MWY0NyIsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDprdWJlLXN5c3RlbTprdWJlcm5ldGVzLWNsaWVudCJ9.TXWm2Wj2Lk64mAGT2CFjkytPGvYYKMxmygOyEDjEdsyUyF3HYsrqkXLRRU6qzLFdurTE3RYA7jh8gT9Rft3yOQGpnUK-MdIXqT0BS_QPRD9hT9wdV7hjd9YGQpmdk9uZ8dZ40lirpqBb1Le2u6bpaESXaGxZQRuR-vSomsk3hqXFqY-0r7_CMjhFSzaRtePUXtgnB5HW2Vr_sfI79yEASRav5at5bX6zcxkq5Ru_ICKHaEPuqslbviSRW4daXKbArfnQkjML3h4-no8n8Qeg-1zOu17AftOddpdyg7L2oxq8uRvvPvCz2lBoFZorhmZQTFkVdPekhiirKEroGEJXtA"
)


func main() {

	client := kubesys.NewKubernetesClient(DefaultMasterUrl, DefaultToken)
	client.Init()

	//createResource(client)
	//getResource(client)
	//updateResource(client)
	deleteResource(client)
	//listResources(client)

	//watchResources(client)
	//watchResource(client)
	//fmt.Println(client.GetKinds())
	//fmt.Println(client.GetFullKinds())
	//fmt.Println(kubesys.ToJsonObject(client.GetKindDesc()).ToString())
}

func watchResource(client *kubesys.KubernetesClient) {
	watcher := kubesys.NewKubernetesWatcher(client, PrintWatchHandler{})
	client.WatchResource("Pod", "default", "busybox", watcher)
}

func watchResources(client *kubesys.KubernetesClient) {
	watcher := kubesys.NewKubernetesWatcher(client, PrintWatchHandler{})
	client.WatchResources("Pod", "", watcher)
}

func createResource(client *kubesys.KubernetesClient) {
	//jsonRes, err := client.CreateResource(createPod())
	_, err := client.CreateResource(createDeployment())

	if err != nil {
		fmt.Println(err)
	}
	//json := kubesys.ToJsonObject(jsonRes)
	//fmt.Println(json.ToString())
}

func deleteResource(client *kubesys.KubernetesClient) {
	//jsonRes, _ := client.DeleteResource("Pod", "default", "busybox")
	jsonRes, _ := client.DeleteResource("Deployment", "default", "kubia")

	json := kubesys.ToJsonObject(jsonRes)
	fmt.Println(json.ToString())
}

func getResource(client *kubesys.KubernetesClient) {
	jsonRes, _ := client.GetResource("Pod", "default", "busybox")
	//fmt.Println(kubesys.ToJsonObject(jsonRes))
	fmt.Println(kubesys.ToGolangMap(jsonRes)["metadata"].(map[string]interface {})["name"].(string))
}

func listResources(client *kubesys.KubernetesClient) {
	jsonRes,_ := client.ListResources("Deployment", "")
	json := kubesys.ToJsonObject(jsonRes)
	fmt.Println(json.ToString())
}

//func createPod() string {
//	return "{\n  \"apiVersion\": \"v1\",\n  \"kind\": \"Pod\",\n  \"metadata\": {\n    \"name\": \"busybox\",\n    \"namespace\": \"default\"\n  },\n  \"spec\": {\n    \"containers\": [\n      {\n        \"image\": \"busybox\",\n        \"env\": [{\n           \"name\": \"abc\",\n           \"value\": \"abc\"\n        }],\n        \"command\": [\n          \"sleep\",\n          \"3600\"\n        ],\n        \"imagePullPolicy\": \"IfNotPresent\",\n        \"name\": \"busybox\"\n      }\n    ],\n    \"restartPolicy\": \"Always\"\n  }\n}"
//}

func createPod() string {
	return "{  \"apiVersion\": \"v1\",  \"kind\": \"Pod\",  \"metadata\": {    \"name\": \"busybox\",    \"namespace\": \"default\"  },  \"spec\": {    \"containers\": [      {        \"image\": \"busybox\",        \"env\": [{           \"name\": \"abc\",           \"value\": \"abc\"        }],        \"command\": [          \"sleep\",          \"3600\"        ],        \"imagePullPolicy\": \"IfNotPresent\",        \"name\": \"busybox\"      }    ],    \"restartPolicy\": \"Always\"  }}"
}

func createDeployment() string {
	return "{\"apiVersion\":\"apps/v1\",\"kind\":\"Deployment\",\"metadata\":{\"annotations\":{},\"name\":\"kubia\",\"namespace\":\"default\"},\"spec\":{\"replicas\":3,\"selector\":{\"matchLabels\":{\"app\":\"kubia\"}},\"template\":{\"metadata\":{\"labels\":{\"app\":\"kubia\"},\"name\":\"kubia\"},\"spec\":{\"containers\":[{\"image\":\"luksa/kubia:v1\",\"name\":\"nodejs\"}]}}}}\n"
}

func updateResource(client *kubesys.KubernetesClient) {

	labels := make(map[string]interface{})
	labels["test"] = "test"

	objRes, _  := client.GetResource("Pod", "default", "busybox")
	obj := kubesys.ToJsonObject(objRes)
	metadata := obj.GetJsonObject("metadata")
	metadata.Put("labels", labels)
	fmt.Println(metadata.ToString())

	obj.Put("metadata", metadata.ToInterface())
	fmt.Println(obj.ToString())

	jsonRes,err := client.UpdateResource(obj.ToString())
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(kubesys.ToJsonObject(jsonRes).ToString())
}

type PrintWatchHandler struct {}

func (p PrintWatchHandler) DoAdded(obj map[string]interface{}) {
	json,_ :=json.Marshal(obj)
	fmt.Println("ADDED: " + string(json))
}
func (p PrintWatchHandler) DoModified(obj map[string]interface{}) {
	json,_ :=json.Marshal(obj)
	fmt.Println("MODIFIED: " + string(json))
}
func (p PrintWatchHandler) DoDeleted(obj map[string]interface{}) {
	json,_ :=json.Marshal(obj)
	fmt.Println("DELETED: " + string(json))
}