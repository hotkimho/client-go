package main

import (
	"context"
	"flag"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

func main() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	watcher, err := client.CoreV1().Pods("default").
		Watch(context.TODO(), metav1.ListOptions{})

	if err != nil {
		klog.Fatal(err)
	}
	// added, modified
	// deleted
	// bookmark
	// error
	// 4가지 5가지 이벤트 타입을 가지는데 이것도 정리하자.
	for event := range watcher.ResultChan() {
		if pod, ok := event.Object.(*corev1.Pod); ok {
			klog.Infof("%s: %s %s", event.Type, pod.GetName())
		}
	}
	fmt.Println("end watch ")
}
