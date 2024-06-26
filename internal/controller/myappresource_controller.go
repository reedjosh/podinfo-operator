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
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
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

// MyAppResources.
//+kubebuilder:rbac:groups=podinfo.podinfo.com,resources=myappresources,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=podinfo.podinfo.com,resources=myappresources/status,verbs=get;update;patch

// K8s Deployments.
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get

// Services.
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services/status,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *MyAppResourceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (res ctrl.Result, retErr error) {
	log := log.FromContext(ctx)
	// Fetch input myApp custom resource.
	myApp := &podinfov1alpha1.MyAppResource{}
	if err := r.Get(ctx, req.NamespacedName, myApp); err != nil {
		// Ignore not-found errors, since it can't be fixed by an immediate requeue (need to wait for a new notification).
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	log.V(1).Info("myappresource found", "name", myApp.Name)

	// If deletion timestamp. Do nothing.
	// If cluster external resources were created, one would need to use a finalizer and perform cleanup in a
	// reconcilDelete function of sorts.
	// Since everything created here is a k8s construct, just let the owner-refs perform cleanup.
	if myApp.GetDeletionTimestamp() != nil {
		return ctrl.Result{}, nil
	}
	// otherwise reconcile.
	return r.reconcile(ctx, req, myApp)
}

// reconcile attempts to create or update a myApp resource per the desired spec.
func (r *MyAppResourceReconciler) reconcile(
	ctx context.Context, req ctrl.Request, myApp *podinfov1alpha1.MyAppResource,
) (ctrl.Result, error) {
	// Create or Updtate deployment and services as needed.
	if err := r.createOrUpdateDeployment(ctx, req, myApp); err != nil {
		return ctrl.Result{}, err
	} else if err = r.createOrUpdateService(ctx, req, myApp); err != nil {
		return ctrl.Result{}, err
	} else if err = r.reconcileRedis(ctx, myApp); err != nil {
		return ctrl.Result{}, err
	}

	// Requeue until the status goes ready.
	// TODO (reedjosh) Should do this with watches instead!
	return ctrl.Result{Requeue: !myApp.Status.Ready}, nil
}

// createOrUpdateDeployment attempts to create or update desired myApp deployment.
// TODO (reedjosh) would use ctrl.CreateOrUpdate but cuases test failures.
func (r *MyAppResourceReconciler) createOrUpdateDeployment(
	ctx context.Context, _ ctrl.Request, myApp *podinfov1alpha1.MyAppResource,
) error {
	log := log.FromContext(ctx)

	// Fetch existing deployment...
	foundDeployment := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: myApp.Name, Namespace: myApp.Namespace}, foundDeployment)

	// Return if err and not just because the deployment wasn't found.
	if err != nil && !k8serrs.IsNotFound(err) {
		return err
	}

	// Deployment not found, create it.
	if err != nil {
		log.V(1).Info("Creating Deployment", "deployment", myApp.Name)
		return r.Create(ctx, buildDeployment(myApp))
	}

	// Deployment found, propagate status.
	// TODO (reedjosh) build a better status rollup of all resources along with a better watch on said resources.
	// TODO (reedjosh) patch the status diff instead of the whole object.
	myApp.Status.Ready = foundDeployment.Status.ReadyReplicas == *myApp.Spec.ReplicaCount
	if err = r.Status().Update(ctx, myApp); err != nil {
		return fmt.Errorf("error patching myappresource: %w", err)
	}

	// Deployment found, update it.
	// TODO (reedjosh) do a nice comparison. Skip update if no diff. Potentially patch instead of update.
	log.V(1).Info("Updating Deployment", "deployment", myApp.Name)
	err = r.Update(ctx, buildDeployment(myApp))
	return err
}

// createOrUpdateService attempts to create or update desired myApp service.
// TODO (reedjosh) would use ctrl.CreateOrUpdate but cuases test failures.
func (r *MyAppResourceReconciler) createOrUpdateService(
	ctx context.Context, _ ctrl.Request, myApp *podinfov1alpha1.MyAppResource,
) error {
	log := log.FromContext(ctx)
	// Fetch existing service...
	foundService := &corev1.Service{}
	err := r.Get(ctx, types.NamespacedName{Name: myApp.Name, Namespace: myApp.Namespace}, foundService)

	// Return if err and not just because the service wasn't found.
	if err != nil && !k8serrs.IsNotFound(err) {
		return err
	}

	// Service not found, create it.
	desiredSvc := buildService(myApp)
	if err != nil {
		log.V(1).Info("Creating Service", "service", myApp.Name)
		return r.Create(ctx, desiredSvc)
	}

	// Service found, update it.
	// TODO (reedjosh) do a nice comparison and even potentially patch instead of update.
	log.V(1).Info("Updating Service", "service", desiredSvc.Name)
	err = r.Update(ctx, desiredSvc)
	return err
}

// SetupWithManager sets up the controller with the Manager.
func (r *MyAppResourceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&podinfov1alpha1.MyAppResource{}).
		Complete(r)
}

// reconcileRedis calls create update or delete for the redis application.
func (r *MyAppResourceReconciler) reconcileRedis(ctx context.Context, myApp *podinfov1alpha1.MyAppResource) error {
	if myApp.Spec.Redis.Enabled {
		if err := r.createOrUpdateRedisDeployment(ctx, myApp); err != nil {
			return err
		}
		return r.createOrUpdateRedisService(ctx, myApp)
	}
	return r.reconcileDeleteRedis(ctx, myApp)
}

func (r *MyAppResourceReconciler) createOrUpdateRedisService(
	ctx context.Context, myApp *podinfov1alpha1.MyAppResource,
) error {
	log := log.FromContext(ctx)
	// Fetch existing service...
	foundService := &corev1.Service{}
	err := r.Get(ctx, types.NamespacedName{Name: myApp.Name + "-redis", Namespace: myApp.Namespace}, foundService)

	// Return if err and not just because the service wasn't found.
	if err != nil && !k8serrs.IsNotFound(err) {
		return err
	}

	// Service not found, create it.
	desiredSvc := buildRedisService(myApp)
	if err != nil {
		log.V(1).Info("Creating Redis Service", "service", myApp.Name)
		if err = r.Create(ctx, desiredSvc); err != nil {
			return err
		}
	}

	// Service found, update it.
	// TODO (reedjosh) do a nice comparison and even potentially patch instead of update.
	log.V(1).Info("Updating Redis Service", "service", desiredSvc.Name)
	err = r.Update(ctx, desiredSvc)
	return err
}

func (r *MyAppResourceReconciler) createOrUpdateRedisDeployment(
	ctx context.Context, myApp *podinfov1alpha1.MyAppResource,
) error {
	log := log.FromContext(ctx)
	// Fetch existing deployment...
	foundDep := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: myApp.Name + redisNamePostfix, Namespace: myApp.Namespace}, foundDep)

	// Return if err and not just because the deployment wasn't found.
	if err != nil && !k8serrs.IsNotFound(err) {
		return err
	}

	// Deployment not found, create it.
	desiredDep := buildRedisDeployment(myApp)
	if err != nil {
		log.V(1).Info("Creating Redis Deployment", "deployment", myApp.Name+redisNamePostfix)
		if err = r.Create(ctx, desiredDep); err != nil {
			return err
		}
	}

	// Deployment found, update it.
	// TODO (reedjosh) do a nice comparison and even potentially patch instead of update.
	log.V(1).Info("Updating Redis Deployment", "name", desiredDep.Name+redisNamePostfix)
	err = r.Update(ctx, desiredDep)
	return err
}

// reconcileDeleteRedis is necesarry to remove the redis deployment on disablement -- not deletion of the myappresource.
func (r *MyAppResourceReconciler) reconcileDeleteRedis(
	ctx context.Context, myApp *podinfov1alpha1.MyAppResource,
) error {
	log := log.FromContext(ctx)
	log.V(1).Info("Deleting Redis components")
	dep := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: myApp.Name + redisNamePostfix, Namespace: myApp.Namespace}}
	if err := r.Delete(ctx, dep); err != nil && !k8serrs.IsNotFound(err) {
		return err
	}
	svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: myApp.Name + redisNamePostfix, Namespace: myApp.Namespace}}
	if err := r.Delete(ctx, svc); err != nil && !k8serrs.IsNotFound(err) {
		return err
	}
	return nil
}
