// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package controllers

import (
	"context"
	"log"

	k8spluginv1alpha1 "gitlab.com/project-emco/core/emco-base/src/monitor/pkg/apis/k8splugin/v1alpha1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	slog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// ResourceBundleStateReconciler reconciles a ResourceBundleState object
type ResourceBundleStateReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=k8splugin.io,resources=resourcebundlestates,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=k8splugin.io,resources=resourcebundlestates/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=k8splugin.io,resources=resourcebundlestates/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ResourceBundleState object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile

func (r *ResourceBundleStateReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	_ = slog.FromContext(ctx)

	log.Println("Reconcile CR", req)

	rbstate := &k8spluginv1alpha1.ResourceBundleState{}

	err := r.Get(context.TODO(), req.NamespacedName, rbstate)

	if err != nil {

		if k8serrors.IsNotFound(err) {

			// CR Deleted, delete status from Git
			appName := req.NamespacedName.Name
			if GitClient != (GitAccessClient{}) {
				err := GitClient.DeleteStatusFromGit(appName)

				return ctrl.Result{}, err
			}

		}

		return ctrl.Result{}, err

	}

	return ctrl.Result{}, nil

}

// SetupWithManager sets up the controller with the Manager.
func (r *ResourceBundleStateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	pred := predicate.GenerationChangedPredicate{}
	return ctrl.NewControllerManagedBy(mgr).
		For(&k8spluginv1alpha1.ResourceBundleState{}).WithEventFilter(pred).
		Complete(r)
}
