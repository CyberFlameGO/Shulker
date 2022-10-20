/*
Copyright 2022.

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
	"fmt"
	"log"
	"os"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
)

func main() {
	namespace := os.Getenv("POD_NAMESPACE")
	name := os.Getenv("POD_NAME")
	if namespace == "" || name == "" {
		log.Fatal(fmt.Errorf("POD_NAMESPACE or POD_NAME is empty"), "invalid configuration")
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal(err, "failed to load kubeconfig in cluster")
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err, "falied to create clientset")
	}

	watch, err := clientset.CoreV1().Pods(namespace).Watch(context.TODO(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.kind=Pod,involvedObject.name=%s", name),
	})
	if err != nil {
		log.Fatal(err.Error(), "failed to setup pod watcher")
	}

	go func() {
		for event := range watch.ResultChan() {
			fmt.Printf("Type: %v\n", event.Type)
			p, ok := event.Object.(*v1.Pod)
			if !ok {
				log.Fatal("unexpected type")
			}
			fmt.Println(p.Status.ContainerStatuses)
			fmt.Println(p.Status.Phase)
		}
	}()

	time.Sleep(5 * time.Second)
}
