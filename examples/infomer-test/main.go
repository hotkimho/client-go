package main

import (
	"context"
	"flag"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

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

	factory := informers.NewSharedInformerFactory(client, 10*time.Second)

	podsInformer := factory.Core().V1().Pods().Informer()
	podsInformer.AddEventHandler(eventHandler)
	factory.Start(ctx.Done())
	factory.WaitForCacheSync(ctx.Done())
	<-ctx.Done()
}

func print(event watch.EventType, pod *corev1.Pod) {
	klog.Infof("%s: %s", event, pod.GetName())
}
