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
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	podinfov1alpha1 "podinfo-operator.com/m/v2/api/v1alpha1"
)

// MyAppResourceReconciler reconciles a MyAppResource object
type MyAppResourceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// Service -- near constant.
func buildService(myApp podinfov1alpha1.MyAppResource) *corev1.Service {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: myApp.Name,
			Namespace: "default",
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


// MyAppResources.
//+kubebuilder:rbac:groups=podinfo.podinfo.com,resources=myappresources,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=podinfo.podinfo.com,resources=myappresources/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=podinfo.podinfo.com,resources=myappresources/finalizers,verbs=update

// K8s Deployments.
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get
//+kubebuilder:rbac:groups=apps,resources=deployments/finalizers,verbs=update

// Services.
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services/status,verbs=get;update
//+kubebuilder:rbac:groups="",resources=services/finalizers,verbs=update


// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MyAppResource object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
func (r *MyAppResourceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (res ctrl.Result, retErr error) {
	log := log.FromContext(ctx)

	// Fetch basic resource.
	var myApp podinfov1alpha1.MyAppResource
	if err := r.Get(ctx, req.NamespacedName, &myApp); err != nil {
		// Ignore not-found errors, since it can't be fixed by an immediate
		// requeue (need to wait for a new notification)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	log.V(1).Info("myApp resource found", "MyAppResource", myApp)

	// if deletion timestamp.
	if myApp.GetDeletionTimestamp() != nil { // Try to delete.
		return r.reconcileDelete(ctx, req, myApp)
	}
	return r.reconcile(ctx, req, myApp)
}

// reconcile attempts to create or update a myApp resource per the desired spec.
func (r *MyAppResourceReconciler) reconcile(
	ctx context.Context, req ctrl.Request, myApp podinfov1alpha1.MyAppResource,
) (ctrl.Result, error) {

	// Create or Updtate deployment and services as needed.
	if res, err := r.createOrUpdateDeployment(ctx, req, myApp); err != nil || res.Requeue {
		return res, err
	}
	// if res, err := r.createOrUpdateService(ctx, req, myApp); err != nil || res.Requeue {
	// 	return res, err
	// }
	
	return ctrl.Result{}, nil
}

func (r *MyAppResourceReconciler) createOrUpdateDeployment(
	ctx context.Context, _ ctrl.Request, myApp podinfov1alpha1.MyAppResource,
) (ctrl.Result, error){
	log := log.FromContext(ctx)

	// Fetch existing deployment...
	foundDeployment := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: myApp.Name, Namespace: myApp.Namespace}, foundDeployment)

	// Return if err and not just because the deployment wasn't found.
	if err != nil && !k8serrs.IsNotFound(err) {
		return ctrl.Result{}, err
	}

	// Deployment not found, create it.
	if err != nil {
		log.V(1).Info("Creating Deployment", "deployment", myApp.Name)
		if err = r.Create(ctx, myApp.AsDeployment()); err != nil {
			return ctrl.Result{},  err
		}
	}

	// Deployment found, update it.
	// TODO (reedjosh) do a nice comparison and even potentially patch instead of update.
	log.V(1).Info("Updating Deployment", "deployment", myApp.Name)
	err = r.Update(ctx, myApp.AsDeployment())
	return ctrl.Result{}, err
}


// func (r *MyAppResourceReconciler) createOrUpdateService(
// 	ctx context.Context, _ ctrl.Request, myApp podinfov1alpha1.MyAppResource,
// ) (ctrl.Result, error){
// 	log := log.FromContext(ctx)
//
// 	// Fetch existing service...
// 	foundService := &corev1.Service{}
// 	err := r.Get(ctx, types.NamespacedName{Name: myApp.Name, Namespace: myApp.Namespace}, foundService)
//
// 	// Return if err and not just because the service wasn't found.
// 	if err != nil && !k8serrs.IsNotFound(err) {
// 		return ctrl.Result{}, err
// 	}
//
// 	// Service not found, create it.
// 	svc := buildService(myApp)
// 	if err != nil {
// 		log.V(1).Info("Creating Service", "service", myApp.Name)
// 		if err = r.Create(ctx, svc); err != nil {
// 			return ctrl.Result{},  err
// 		}
// 	}
//
// 	// Service found, update it.
// 	// TODO (reedjosh) do a nice comparison and even potentially patch instead of update.
// 	log.V(1).Info("Updating Service", "service", svc.Name)
// 	err = r.Update(ctx, svc)
// 	return ctrl.Result{}, err
// }


// reconcileDelete attempts to delete a myApp resource with a deletion timestamp.
func (r *MyAppResourceReconciler) reconcileDelete(
	ctx context.Context, _ ctrl.Request, myApp podinfov1alpha1.MyAppResource,
) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Fetch existing deployments if existing...
	foundDeployment := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: myApp.Name, Namespace: myApp.Namespace}, foundDeployment)

	// Return if err and not just because the deployment wasn't found.
	if err != nil && !k8serrs.IsNotFound(err) {
		return ctrl.Result{}, err
	}

	// Object didn't exist, so do nothing.
	if err != nil {
		return ctrl.Result{}, nil
	}

	// Found matching deployment -- delete it.
	log.V(1).Info("Deleting Deployment", "deployment", myApp.Name)
	return ctrl.Result{}, r.Delete(ctx, foundDeployment)
}


// SetupWithManager sets up the controller with the Manager.
func (r *MyAppResourceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&podinfov1alpha1.MyAppResource{}).
		Complete(r)
}
