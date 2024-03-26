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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/resource"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	podinfov1alpha1 "podinfo-operator.com/m/v2/api/v1alpha1"
)

var _ = Describe("MyAppResource Controller Support Functions", func() {
	const resourceName = "test-resource"

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

	BeforeEach(func() {})
	AfterEach(func() {})

	It("should successfully convert a myappresource to a deployment", func() {
		d := buildDeployment(myappresource)
		Expect(d.Spec.Template.Spec.Containers[0].Name).To(Equal("podinfo"))
		Expect(d.Name).To(Equal(myappresource.Name))
		Expect(d.Namespace).To(Equal(myappresource.Namespace))
	})

	It("should successfully build a matching service", func() {
		d := buildService(myappresource)
		Expect(d.Spec.Ports[0].Port).To(Equal(int32(9898)))
	})
})
