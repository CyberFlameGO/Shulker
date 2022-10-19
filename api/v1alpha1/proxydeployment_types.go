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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ProxyDeploymentSpec defines the desired state of ProxyDeployment
type ProxyDeploymentSpec struct {
	// Reference to a MinecraftCluster. Adding this will enroll
	// this ProxyDeployment to be part of a MinecraftCluster.
	//+kubebuilder:validation:Required
	ClusterRef MinecraftClusterRef `json:"clusterRef,omitempty"`

	// Number of Proxy replicas to create.
	//+kubebuilder:validation:Required
	Replicas int32 `json:"replicas,omitempty"`

	// The desired state of the Kubernetes Service to create for the
	// Proxy Deployment.
	//+kubebuilder:validation:Required
	Service ProxyDeploymentServiceSpec `json:"service,omitempty"`

	// Template defining the content of the created Proxies.
	//+kubebuilder:validation:Required
	Template ProxyTemplate `json:"template,omitempty"`
}

// Configuration attributes for the Service resource.
type ProxyDeploymentServiceSpec struct {
	// Type of Service to create. Must be one of: ClusterIP, LoadBalancer, NodePort.
	//+kubebuilder:validation:Enum=ClusterIP;LoadBalancer;NodePort
	//+kubebuilder:default="LoadBalancer"
	Type corev1.ServiceType `json:"type,omitempty"`

	// Annotations to add to the Service.
	Annotations map[string]string `json:"annotations,omitempty"`

	// Describe how nodes distribute service traffic to the proxy.
	//+kubebuilder:validation:Enum=Cluster;Local
	//+kubebuilder:default="Cluster"
	ExternalTrafficPolicy corev1.ServiceExternalTrafficPolicyType `json:"externalTrafficPolicy,omitempty"`
}

type ProxyDeploymentStatusCondition string

const (
	ProxyDeploymentAvailableCondition ProxyDeploymentStatusCondition = "Available"
)

// ProxyDeploymentStatus defines the observed state of ProxyDeployment
type ProxyDeploymentStatus struct {
	// Conditions represent the latest available observations of a
	// ProxyDeployment object.
	// Known .status.conditions.type are: "Available".
	//+kubebuilder:validation:Required
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`

	// Number of total replicas in this ProxyDeployment.
	Replicas int32 `json:"replicas"`

	// Number of available replicas in this ProxyDeployment.
	AvailableReplicas int32 `json:"availableReplicas"`

	// Number of unavailable replicas in this ProxyDeployment.
	UnavailableReplicas int32 `json:"unavailableReplicas"`

	// Pod label selector.
	Selector string `json:"selector"`
}

func (s *ProxyDeploymentStatus) SetCondition(condition ProxyDeploymentStatusCondition, status metav1.ConditionStatus, reason string, message string) metav1.Condition {
	c := metav1.Condition{
		Type:    string(condition),
		Status:  status,
		Reason:  reason,
		Message: message,
	}

	meta.SetStatusCondition(&s.Conditions, c)
	return c
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.replicas,selectorpath=.status.selector
//+kubebuilder:printcolumn:name="Replicas",type="integer",JSONPath=".status.replicas"
//+kubebuilder:printcolumn:name="Available Replicas",type="integer",JSONPath=".status.availableReplicas"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:resource:shortName={"skrpd"},categories=all

// ProxyDeployment is the Schema for the proxydeployments API
type ProxyDeployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProxyDeploymentSpec   `json:"spec,omitempty"`
	Status ProxyDeploymentStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ProxyDeploymentList contains a list of ProxyDeployment
type ProxyDeploymentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ProxyDeployment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ProxyDeployment{}, &ProxyDeploymentList{})
}
