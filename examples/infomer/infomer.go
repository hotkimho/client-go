package main

import (
	"context"
	"flag"
	"fmt"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
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

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	lw := &cache.ListWatch{
		ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
			return client.CoreV1().Pods(metav1.NamespaceDefault).List(ctx, opts)
		},
		WatchFunc: func(opts metav1.ListOptions) (watch.Interface, error) {
			return client.CoreV1().Pods(metav1.NamespaceDefault).Watch(ctx, opts)
		},
	}

	eventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			print(watch.Added, obj.(*corev1.Pod))
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			old := oldObj.(*corev1.Pod)
			new := newObj.(*corev1.Pod)
			if old.ResourceVersion == new.ResourceVersion {
				return
			}
			print(watch.Modified, new)
		},
		DeleteFunc: func(obj interface{}) {
			print(watch.Deleted, obj.(*corev1.Pod))
		},
	}

	fmt.Println("start informer")
	store, informer := cache.NewInformer(lw, &corev1.Pod{}, 30*time.Second, eventHandler)

	go informer.Run(ctx.Done())

	fmt.Println("1121212121212121212121")
	time.Sleep(5 * time.Second)
	aa := store.ListKeys()
	for _, a := range aa {
		fmt.Println("store keys : ", a)
	}

}
