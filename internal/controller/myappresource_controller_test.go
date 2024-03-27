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

package controller

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	podinfov1alpha1 "podinfo-operator.com/m/v2/api/v1alpha1"
)

func ptr[T any](v T) *T { return &v }

// performReconcilation triggers one cycle of the MyAppResource reconciler.
func performReconcilation(ctx context.Context, namespacedName types.NamespacedName) {
	controllerReconciler := &MyAppResourceReconciler{Client: k8sClient, Scheme: k8sClient.Scheme()}
	_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
	Expect(err).NotTo(HaveOccurred())
}

var _ = Describe("MyAppResource Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()
		myappresource := &podinfov1alpha1.MyAppResource{}

		namespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}

		BeforeEach(func() {
			// Rest the myappresource.
			*myappresource = podinfov1alpha1.MyAppResource{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: "default",
				},
				Spec: podinfov1alpha1.MyAppResourceSpec{
					ReplicaCount: ptr(int32(3)),
					Image: podinfov1alpha1.Image{
						Repository: "ghcr.io/stefanprodan/podinfo",
						Tag:        "latest",
					},
					Resources: podinfov1alpha1.Resources{
						CPURequest:  *resource.NewQuantity(100, "m"),
						MemoryLimit: *resource.NewQuantity(64, "mi"),
					},
					UI: podinfov1alpha1.UI{
						Color:   "#34577c",
						Message: "some string",
					},
				},
			}

			By("creating the custom resource for the Kind MyAppResource")
			err := k8sClient.Get(ctx, namespacedName, myappresource)
			if err != nil && errors.IsNotFound(err) {
				Expect(k8sClient.Create(ctx, myappresource)).To(Succeed())
			}
		})

		AfterEach(func() {
			resource := &podinfov1alpha1.MyAppResource{}
			Expect(k8sClient.Get(ctx, namespacedName, resource)).To(Succeed())

			By("Cleanup the specific resource instance MyAppResource")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})

		It("should successfully reconcile the myappresource", func() {
			By("creating the deployment without redis, and the correct spec when not redis enabled in the myappresource spec")
			performReconcilation(ctx, namespacedName)
			deployment := &appsv1.Deployment{}
			svc := &corev1.Service{}
			Expect(k8sClient.Get(ctx, namespacedName, deployment)).To(Succeed())
			Expect(k8sClient.Get(ctx, namespacedName, svc)).To(Succeed())
			Expect(len(deployment.Spec.Template.Spec.Containers)).To(Equal(1))
			Expect(deployment.Spec.Template.Spec.Containers[0].Env).To(ContainElements(
				[]corev1.EnvVar{
					{Name: "PODINFO_UI_COLOR", Value: myappresource.Spec.UI.Color},
					{Name: "PODINFO_UI_MESSAGE", Value: myappresource.Spec.UI.Message},
				},
			))

			By("ensuring owner-refs are set on each created resource to ensure cleanup")
			Expect(k8sClient.Get(ctx, namespacedName, myappresource)).To(Succeed()) // Needed for UID.
			ownerRef := []metav1.OwnerReference{{
				APIVersion:         "podinfo.podinfo.com/v1alpha1",
				BlockOwnerDeletion: ptr(true),
				Controller:         ptr(true),
				Kind:               "MyAppResource",
				Name:               "resourceName",
				UID:                myappresource.UID,
			}}
			Expect(deployment.OwnerReferences[0]).To(Equal(ownerRef))
			Expect(svc.OwnerReferences[0]).To(Equal(ownerRef))

			By("updating the deployment with redis enabled when the myappresource is set to Redis enabled.")
			myappresource.Spec.Redis.Enabled = true
			Expect(k8sClient.Update(ctx, myappresource)).To(Succeed())
			performReconcilation(ctx, namespacedName)

			deployment = &appsv1.Deployment{}
			Expect(k8sClient.Get(ctx, namespacedName, deployment)).Should(Succeed())
			Expect(deployment.Spec.Replicas).To(Equal(myappresource.Spec.ReplicaCount))

			Expect(deployment.Spec.Template.Spec.Containers[1].Name).To(Equal("redis"))
			Expect(deployment.Spec.Template.Spec.Containers[0].Env).To(ContainElements(
				[]corev1.EnvVar{{Name: "PODINFO_CACHE_SERVER", Value: "tcp://localhost:6379"}},
			))
		})
	})
})
