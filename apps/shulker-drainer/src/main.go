/*
Copyright (c) Jérémy Levilain
SPDX-License-Identifier: GPL-3.0-or-later
*/

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"sync"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	kubernetes "k8s.io/client-go/dynamic"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"

	shulkermciov1alpha1 "github.com/iamblueslime/shulker/libs/crds/v1alpha1"
)

const drainFileName = "no-drain"

func createNoDrainFile(drainLockDir string) error {
	file := path.Join(drainLockDir, drainFileName)
	content := []byte{}
	return os.WriteFile(file, content, 0644)
}

func drain(drainLockDir string) error {
	file := path.Join(drainLockDir, drainFileName)
	if _, err := os.Stat(file); err == nil {
		return os.Remove(file)
	}
	return nil
}

func main() {
	namespace := os.Getenv("POD_NAMESPACE")
	name := os.Getenv("POD_NAME")
	drainLockDir := os.Getenv("SHULKER_DRAIN_LOCK_DIR")
	if namespace == "" || name == "" || drainLockDir == "" {
		log.Fatal(fmt.Errorf("POD_NAMESPACE, POD_NAME or SHULKER_DRAIN_LOCK_DIR is empty"), "invalid configuration")
	}

	err := createNoDrainFile(drainLockDir)
	if err != nil {
		log.Fatal(err, "failed to create no-drain file")
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal(err, "failed to load kubeconfig in cluster")
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err, "falied to create clientset")
	}

	watcher, err := clientset.Resource(schema.GroupVersionResource{
		Group:    shulkermciov1alpha1.GroupVersion.Group,
		Version:  shulkermciov1alpha1.GroupVersion.Version,
		Resource: "proxies",
	}).Namespace(namespace).Watch(context.Background(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", name),
	})
	if err != nil {
		log.Fatal(err, "failed to setup proxy watcher")
	}

	var lock sync.Mutex
	log.Println("watching proxy")

	go func() {
		lock.Lock()

		for event := range watcher.ResultChan() {
			proxy := event.Object.(*unstructured.Unstructured)

			if event.Type == watch.Modified {
				log.Println("proxy was modified")

				if value, ok := proxy.GetAnnotations()[shulkermciov1alpha1.ProxyDrainAnnotationName]; ok && value == "true" {
					log.Println("draining annotation found")
					err := drain(drainLockDir)
					if err != nil {
						log.Fatal(err, "failed to drain")
					}
					lock.Unlock()
				}
			}
		}
	}()

	time.Sleep(3 * time.Second)
	lock.Lock()
	log.Println("exiting")
}
