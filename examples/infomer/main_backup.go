package main

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"
	"time"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

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

	var lastResourceVersion string

	podWatcher, err := client.CoreV1().Pods("default").
		Watch(context.TODO(), metav1.ListOptions{
			AllowWatchBookmarks: true,
			ResourceVersion:     lastResourceVersion,
		})

	if err != nil {
		klog.Fatal(err)
	}

	go func() {
		for event := range podWatcher.ResultChan() {
			// 북마크 이벤트를 수신하면 리소스버전 갱신
			if event.Type == "BOOKMARK" {
				klog.Infof("[POD] BOOKMARK111111111: %s", event.Object.(*corev1.Pod).GetResourceVersion())
				lastResourceVersion = event.Object.(*corev1.Pod).GetResourceVersion()
			} else if pod, ok := event.Object.(*corev1.Pod); ok {
				klog.Infof("[POD] %s: %s %s", event.Type, pod.GetName(), pod.GetResourceVersion())
				// 이벤트 수신 시마다 리소스버전 갱신
				//lastResourceVersion = pod.GetResourceVersion()
			}
		}
	}()

	time.Sleep(1200 * time.Second)
	fmt.Println("stop pod watcher")
	podWatcher.Stop()

	/*
		이 시간에 새로운 파드를 생성/삭제
	*/

	time.Sleep(10 * time.Second)

	fmt.Println("restart pod watcher")
	fmt.Println(lastResourceVersion)
	// 북마크 이벤트 간격은 10분(default)
	podWatcher, err = client.CoreV1().Pods("default").
		Watch(context.TODO(), metav1.ListOptions{
			ResourceVersion:     lastResourceVersion,
			AllowWatchBookmarks: true,
		})
	if err != nil {
		klog.Fatal(err)
	}

	for event := range podWatcher.ResultChan() {
		if event.Type == "BOOKMARK" {
			klog.Infof("[POD] BOOKMARK22222: %s", event.Object.(*corev1.Pod).GetResourceVersion())
			lastResourceVersion = event.Object.(*corev1.Pod).GetResourceVersion()
		} else if pod, ok := event.Object.(*corev1.Pod); ok {
			klog.Infof("[POD] %s: %s", event.Type, pod.GetName())
			lastResourceVersion = pod.GetResourceVersion()
		}
	}

	time.Sleep(60 * time.Second)
	// close watcher
	podWatcher.Stop()
}
