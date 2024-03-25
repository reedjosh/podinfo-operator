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
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	podinfov1alpha1 "podinfo-operator.com/m/v2/api/v1alpha1"
)

func ptr[T any](v T) *T { return &v }

var _ = Describe("MyAppResource Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default", // TODO(user):Modify as needed
		}

		myappresource := &podinfov1alpha1.MyAppResource{
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
				Redis: podinfov1alpha1.Redis{Enabled: true},
			},
		}

		BeforeEach(func() {
			By("creating the custom resource for the Kind MyAppResource")
			err := k8sClient.Get(ctx, typeNamespacedName, myappresource)
			if err != nil && errors.IsNotFound(err) {
				Expect(k8sClient.Create(ctx, myappresource)).To(Succeed())
			}
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			resource := &podinfov1alpha1.MyAppResource{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance MyAppResource")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})

		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &MyAppResourceReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			// TODO(user): Add more specific assertions depending on your controller's reconciliation logic.
			// Example: If you expect a certain status condition after reconciliation, verify it here.
		})
	})
})
