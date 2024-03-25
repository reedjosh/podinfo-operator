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
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	MyAppResourceFinalizaer = "myappresource.podinfo.podinfo.com"
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

// AsDeployment converts a MyAppResourceSpec to a k8s Deployment Spec.
func (myApp *MyAppResource) AsDeployment() *appsv1.Deployment {
	dep := &appsv1.Deployment{}

	// TODO: (reedjosh) use a better labeling scheme.
	dep.Labels = map[string]string{"app.kubernetes.io/name": myApp.Name}
	dep.Spec.Selector = &metav1.LabelSelector{MatchLabels: map[string]string{"app.kubernetes.io/name": myApp.Name}}
	dep.Spec.Template.Labels = map[string]string{"app.kubernetes.io/name": myApp.Name}
	dep.Namespace = "default"
	dep.Name = myApp.Name

	dep.Spec.Replicas = myApp.Spec.ReplicaCount

	// Assumes containers[0] is always popdinfo container and [1] is always redis if existing.
	dep.Spec.Template.Spec.Containers = []corev1.Container{
		{
			Name:  "podinfo",
			Image: fmt.Sprintf("%s:%s", myApp.Spec.Image.Repository, myApp.Spec.Image.Tag),
			Resources: corev1.ResourceRequirements{
				Limits:   corev1.ResourceList{corev1.ResourceMemory: myApp.Spec.Resources.MemoryLimit},
				Requests: corev1.ResourceList{corev1.ResourceCPU: myApp.Spec.Resources.CPURequest},
			},
			Env: []corev1.EnvVar{
				{Name: "PODINFO_UI_COLOR", Value: myApp.Spec.UI.Color},
				{Name: "PODINFO_UI_MESSAGE", Value: myApp.Spec.UI.Message},
			},
			Ports: []corev1.ContainerPort{
				{ContainerPort: 9898, Name: "http"},
				{ContainerPort: 9797, Name: "http-metrics"},
				{ContainerPort: 9999, Name: "grpc"},
			},
		},
	}

	// If Redis is enabled, add additional container to deployment and set env var as such.
	if myApp.Spec.Redis.Enabled {
		dep.Spec.Template.Spec.Containers = append(dep.Spec.Template.Spec.Containers,
			corev1.Container{
				Name:  "redis",
				Image: "redis:alpine3.19",
				Resources: corev1.ResourceRequirements{
					Limits:   corev1.ResourceList{corev1.ResourceMemory: myApp.Spec.Redis.Resources.MemoryLimit},
					Requests: corev1.ResourceList{corev1.ResourceCPU: myApp.Spec.Redis.Resources.CPURequest},
				},
			},
		)
		dep.Spec.Template.Spec.Containers[0].Env = append(
			dep.Spec.Template.Spec.Containers[0].Env,
			corev1.EnvVar{Name: "PODINFO_CACHE_SERVER", Value: "tcp://localhost:6379"},
		)
	}

	return dep
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
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
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
