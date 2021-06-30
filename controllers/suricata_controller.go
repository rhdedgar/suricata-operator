/*
Copyright 2021 Doug Edgar.

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

package controllers

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appsv1 "k8s.io/api/apps/v1"

	"github.com/pingcap/errors"
	managedv1alpha1 "github.com/rhdedgar/suricata-operator/api/v1alpha1"
	"github.com/rhdedgar/suricata-operator/k8s"
)

// SuricataReconciler reconciles a Suricata object
type SuricataReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=managed.openshift.io,namespace="openshift-suricata-operator",resources=suricatas,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=managed.openshift.io,namespace="openshift-suricata-operator",resources=suricatas/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=managed.openshift.io,namespace="openshift-suricata-operator",resources=suricatas/finalizers,verbs=update

// +kubebuilder:rbac:groups=security.openshift.io,namespace="openshift-suricata-operator",resources=securitycontextconstraints,verbs=use
// +kubebuilder:rbac:groups=security.openshift.io,namespace="openshift-suricata-operator",resources=securitycontextconstraints,verbs=get;create;update
// +kubebuilder:rbac:groups=security.openshift.io,namespace="openshift-suricata-operator",resourceNames=privileged,resources=securitycontextconstraints,verbs=get;create;update

// +kubebuilder:rbac:groups=core,namespace="openshift-suricata-operator",resources=pods,verbs=get;list;
// +kubebuilder:rbac:groups=core,namespace="openshift-suricata-operator",resources=events,verbs=create;watch;list;patch
// +kubebuilder:rbac:groups=core,namespace="openshift-suricata-operator",resources=secrets;services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,namespace="openshift-suricata-operator",resources=deployments;daemonsets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Suricata object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *SuricataReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// Fetch the Suricata instance
	instance := &managedv1alpha1.Suricata{}
	err := r.Client.Get(context.TODO(), req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Define a new DaemonSet object
	daemonSet := k8s.SuricataDaemonSet(instance)

	// Set Suricata instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, daemonSet, r.Scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this DaemonSet already exists
	found := &appsv1.DaemonSet{}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: daemonSet.Name, Namespace: daemonSet.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		err = r.Client.Create(context.TODO(), daemonSet)
		if err != nil {
			return reconcile.Result{}, err
		}

		// DaemonSet created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SuricataReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&managedv1alpha1.Suricata{}).
		Owns(&managedv1alpha1.Suricata{}).
		Complete(r)
}
