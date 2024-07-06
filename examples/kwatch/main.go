/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()
	if len(os.Args) < 3 || len(os.Args) > 4 {
		panic("Usage: kwatch <apigroup/version> <resource> [namespace]")
	}

	apigroup := os.Args[1]
	resource := os.Args[2]
	namespace := "default"
	if len(os.Args) == 4 {
		namespace = os.Args[3]
	}

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	clientRes := getGroupVersionResource(apigroup, resource)

	watcher, err := client.Resource(clientRes).Namespace(namespace).Watch(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	for event := range watcher.ResultChan() {
		d, ok := event.Object.(*unstructured.Unstructured)
		if !ok {
			fmt.Println("failed to convert to unstructured")
			continue
		}

		fmt.Printf("%s\t%s\t%s\t%s\n", event.Type, d.GetName(), d.GetUID(), d.GetNamespace())
	}

}

func getGroupVersionResource(apigroup, resource string) schema.GroupVersionResource {
	v := "v1"
	g := apigroup

	if strings.Contains(apigroup, "/") {
		ag := strings.Split(apigroup, "/")

		g = ag[0]
		v = ag[1]
	}
	if g == "v1" || g == "core" {
		g = ""
	}

	switch resource {
	case "pod":
		resource = "pods"
	case "deployment", "deploy":
		resource = "deployments"
	case "replicaset", "rs":
		resource = "replicasets"
	case "service", "svc":
		resource = "services"
	case "endpoint", "ep":
		resource = "endpoints"
	case "daemonset", "ds":
		resource = "daemonsets"
	case "job":
		resource = "jobs"
	}
	return schema.GroupVersionResource{Group: g, Version: v, Resource: resource}
}
