/*
Copyright 2024 Joshua Reed.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR  CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"

	corev1 "k8s.io/api/core/v1"

	podinfov1alpha1 "podinfo-operator.com/m/v2/api/v1alpha1"
)

const (
	redisNamePostfix = "-redis"
)

// buildService builds a service for a podinfo deployment.
func buildService(myApp *podinfov1alpha1.MyAppResource) *corev1.Service {
	ownerGVK := schema.GroupVersionKind{
		Group:   "podinfo.podinfo.com",
		Version: "v1alpha1",
		Kind:    "MyAppResource",
	}
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            myApp.Name,
			Namespace:       "default",
			Labels:          map[string]string{podinfov1alpha1.MyAppResourceLabelName: myApp.Name},
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(myApp, ownerGVK)},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{Name: "http", Port: 9898, TargetPort: intstr.FromString("http")},
				{Name: "grpc", Port: 9999, TargetPort: intstr.FromString("grpc")},
			},
			Selector: map[string]string{"app.kubernetes.io/name": myApp.Name},
		},
	}
	return svc
}

// buildDeployment converts a MyAppResourceSpec to a k8s Deployment Spec.
func buildDeployment(myApp *podinfov1alpha1.MyAppResource) *appsv1.Deployment {
	ownerGVK := schema.GroupVersionKind{
		Group:   "podinfo.podinfo.com",
		Version: "v1alpha1",
		Kind:    "MyAppResource",
	}
	dep := &appsv1.Deployment{}

	// TODO: (reedjosh) use a better labeling scheme.
	dep.Name = myApp.Name
	dep.Namespace = "default"
	dep.Labels = map[string]string{"app.kubernetes.io/name": myApp.Name}
	dep.OwnerReferences = []metav1.OwnerReference{*metav1.NewControllerRef(myApp, ownerGVK)}
	dep.Spec.Template.Labels = map[string]string{"app.kubernetes.io/name": myApp.Name}
	dep.Spec.Selector = &metav1.LabelSelector{MatchLabels: map[string]string{"app.kubernetes.io/name": myApp.Name}}
	dep.Spec.Replicas = myApp.Spec.ReplicaCount
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

	// If Redis is enabled, set env var as such.
	if myApp.Spec.Redis.Enabled {
		dep.Spec.Template.Spec.Containers[0].Env = append(
			dep.Spec.Template.Spec.Containers[0].Env,
			corev1.EnvVar{
				Name: "PODINFO_CACHE_SERVER",
				Value: fmt.Sprintf(
					"tcp://%s.%s.svc.cluster.local:6379",
					myApp.Name+redisNamePostfix,
					myApp.Namespace),
			},
		)
	}

	return dep
}

// buildService builds a service for a podinfo deployment.
func buildRedisService(myApp *podinfov1alpha1.MyAppResource) *corev1.Service {
	ownerGVK := schema.GroupVersionKind{
		Group:   "podinfo.podinfo.com",
		Version: "v1alpha1",
		Kind:    "MyAppResource",
	}
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            myApp.Name + redisNamePostfix,
			Namespace:       "default",
			Labels:          map[string]string{podinfov1alpha1.MyAppResourceLabelName: myApp.Name},
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(myApp, ownerGVK)},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{Name: "redis", Port: 6379, TargetPort: intstr.FromString("redis")},
			},
			Selector: map[string]string{"app.kubernetes.io/name": myApp.Name + redisNamePostfix},
		},
	}
	return svc
}

// buildRedisDeployment converts a MyAppResourceSpec to a k8s Deployment Spec.
func buildRedisDeployment(myApp *podinfov1alpha1.MyAppResource) *appsv1.Deployment {
	ownerGVK := schema.GroupVersionKind{Group: "podinfo.podinfo.com", Version: "v1alpha1", Kind: "MyAppResource"}

	// TODO: (reedjosh) use a better labeling scheme.
	dep := &appsv1.Deployment{}
	dep.Name = myApp.Name + redisNamePostfix
	dep.Namespace = "default"
	dep.Labels = map[string]string{"app.kubernetes.io/name": myApp.Name}
	dep.OwnerReferences = []metav1.OwnerReference{*metav1.NewControllerRef(myApp, ownerGVK)}
	dep.Spec.Replicas = myApp.Spec.ReplicaCount
	dep.Spec.Template.Labels = map[string]string{"app.kubernetes.io/name": myApp.Name + redisNamePostfix}
	dep.Spec.Selector = &metav1.LabelSelector{MatchLabels: map[string]string{"app.kubernetes.io/name": myApp.Name + redisNamePostfix}}
	dep.Spec.Template.Spec.Containers = append(dep.Spec.Template.Spec.Containers,
		corev1.Container{
			Name:  "redis",
			Image: "redis:alpine3.19",
			Resources: corev1.ResourceRequirements{
				Limits:   corev1.ResourceList{corev1.ResourceMemory: myApp.Spec.Redis.Resources.MemoryLimit},
				Requests: corev1.ResourceList{corev1.ResourceCPU: myApp.Spec.Redis.Resources.CPURequest},
			},
			Ports: []corev1.ContainerPort{{Name: "redis", ContainerPort: 6379}},
		},
	)
	return dep
}
