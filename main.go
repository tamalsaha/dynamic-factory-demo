package main

import (
	"context"
	"fmt"
	"github.com/tamalsaha/dynamic-demo/factory"
	"k8s.io/apimachinery/pkg/labels"
	"log"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/dynamic/dynamiclister"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main(){
	masterURL := ""
	kubeconfigPath := filepath.Join(homedir.HomeDir(), ".kube", "config")

	config, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfigPath)
	if err != nil {
		log.Fatalf("Could not get Kubernetes config: %s", err)
	}

	dc := dynamic.NewForConfigOrDie(config)

	gvrPod := schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "pods",
	}

	directFactory := factory.New(dc)

	sel, err := labels.Parse("k8s-app=kube-dns")
	if err != nil {
		panic(err)
	}
	objects, err := directFactory.ForResource(gvrPod).List(sel)
	if err != nil {
		panic(err)
	}
	for _, obj := range objects {
		fmt.Println(obj.GetName())
	}

	fmt.Println("---------------------------------------------------------")

	cachedFactory := factory.NewCached(dc, 0, context.TODO().Done())
	objects2, err := cachedFactory.ForResource(gvrPod).List(sel)
	if err != nil {
		panic(err)
	}
	for _, obj := range objects2 {
		fmt.Println(obj.GetName())
	}
}

func main_raw() {
	masterURL := ""
	kubeconfigPath := filepath.Join(homedir.HomeDir(), ".kube", "config")

	config, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfigPath)
	if err != nil {
		log.Fatalf("Could not get Kubernetes config: %s", err)
	}

	dc := dynamic.NewForConfigOrDie(config)

	nodes, err := dc.Resource(schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "nodes",
	}).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	for _, obj := range nodes.Items {
		fmt.Printf("%+v\n", obj.GetName())
	}

	ctx := context.TODO()

	sel, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels:      nil,
		MatchExpressions: nil,
	})
	if err != nil {
		panic(err)
	}

	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(dc, 0, metav1.NamespaceAll, nil)

	gvrPod := schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "pods",
	}
	informerPod := factory.ForResource(gvrPod)
	factory.Start(ctx.Done())
	if synced := factory.WaitForCacheSync(ctx.Done()); !synced[gvrPod] {
		panic(fmt.Sprintf("informer for %s hasn't synced", gvrPod))
	}
	listerPod := dynamiclister.New(informerPod.Informer().GetIndexer(), gvrPod)

	result, err := listerPod.Namespace("default").List(sel)
	if err != nil {
		panic(err)
	}
	for _, obj := range result {
		fmt.Println(obj.GetName())
	}

	gvrDep := schema.GroupVersionResource{
		Group:    "apps",
		Version:  "v1",
		Resource: "deployments",
	}
	informerDep := factory.ForResource(gvrDep)
	factory.Start(ctx.Done())
	if synced := factory.WaitForCacheSync(ctx.Done()); !synced[gvrDep] {
		panic(fmt.Sprintf("informer for %s hasn't synced", gvrDep))
	}
	listerDep := dynamiclister.New(informerDep.Informer().GetIndexer(), gvrDep)
	result2, err := listerDep.Namespace("kube-system").List(sel)
	if err != nil {
		panic(err)
	}
	for _, obj := range result2 {
		fmt.Println(obj.GetName())
	}
}
