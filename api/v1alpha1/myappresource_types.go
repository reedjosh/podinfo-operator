/*
Copyright 2024 Joshua Reed.

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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	MyAppResourceFinalizer = "myappresource.podinfo.podinfo.com"
	MyAppResourceLabelName = "myappresource.podinfo.podinfo.com/name"
)

// MyAppResourceSpec defines the desired state of MyAppResource
type MyAppResourceSpec struct {
	// ReplicaCount is the number of desired replicas of myappresource to launch.
	ReplicaCount *int32 `json:"replicaCount"`

	// Specify the myappresource image to run.
	Image Image `json:"image"`

	// UI spec for User Interface options.
	UI UI `json:"ui,omitempty"`

	Redis Redis `json:"redis,omitempty"`

	// The podinfo deployment resources spec.
	Resources Resources `json:"resources,omitempty" protobuf:"bytes,8,opt,name=resources"`
}

// UI spec for User Interface options.
type UI struct {
	// Color is the UI color desired.
	// +optional
	Color string `json:"color,omitempty"`

	// Color is the UI color desired.
	// +optional
	Message string `json:"message,omitempty"`
}

type Redis struct {
	// Enable or disable redis usage.
	Enabled bool `json:"enabled"`

	// The Redis resources spec.
	Resources Resources `json:"resources,omitempty" protobuf:"bytes,8,opt,name=resources"`
}

type Image struct {
	// Repository is the image to pull.
	Repository string `json:"repository,omitempty"`

	// Tag is the image version to pull
	Tag string `json:"tag"`
}

type Resources struct {
	// memoryLimit is the mem limit for a myappresource dpod.
	// +optional
	MemoryLimit resource.Quantity `json:"memoryLimit,omitempty"`

	// memoryLimit is the mem limit for a myappresource dpod.
	CPURequest resource.Quantity `json:"cpuRequest"`
}

// MyAppResourceStatus defines the observed state of MyAppResource
type MyAppResourceStatus struct {
	// ready indicates whether the podinfo deployment's ready replicas is equal to it's requested replicas.
	Ready bool `json:"ready"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// MyAppResource is the Schema for the myappresources API
type MyAppResource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MyAppResourceSpec   `json:"spec,omitempty"`
	Status MyAppResourceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MyAppResourceList contains a list of MyAppResource
type MyAppResourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MyAppResource `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MyAppResource{}, &MyAppResourceList{})
}
