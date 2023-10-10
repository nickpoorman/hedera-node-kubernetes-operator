/*
Copyright 2023.

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

	apps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	appv1alpha1 "github.com/nickpoorman/hedera-node-kubernetes-operator/api/v1alpha1"
)

const httpEchoImage = "hashicorp/http-echo"

// TenantReconciler reconciles a Tenant object
type TenantReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=github.com/nickpoorman/hedera-node-kubernetes-operator/app,resources=tenants,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=github.com/nickpoorman/hedera-node-kubernetes-operator/app,resources=tenants/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=github.com/nickpoorman/hedera-node-kubernetes-operator/app,resources=tenants/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(nickpoorman): Modify the Reconcile function to compare the state specified by
// the Tenant object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *TenantReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	_ = logger.WithValues("tenant", req.NamespacedName)

	// Get the tenant.
	tenant := &appv1alpha1.Tenant{}
	err := r.Get(ctx, req.NamespacedName, tenant)
	if err != nil {
		if errors.IsNotFound(err) {
			// CR was deleted, cleanup is handled by Kubernetes garbage collection.
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Check if deployment exists, if not create it
	deploymentName := tenant.Spec.Name + "-echo-deployment"
	deployment := &apps.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: deploymentName, Namespace: req.Namespace}, deployment)
	if err != nil && errors.IsNotFound(err) {
		deployment := &apps.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      deploymentName,
				Namespace: req.Namespace,
			},
			Spec: apps.DeploymentSpec{
				Replicas: int32Ptr(1),
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"tenant": tenant.Spec.Name},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{"tenant": tenant.Spec.Name},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "echo",
								Image: httpEchoImage,
								Args:  []string{"-text", "Hello from tenant: " + tenant.Spec.Name},
							},
						},
					},
				},
			},
		}

		if err := ctrl.SetControllerReference(tenant, deployment, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}

		if err = r.Create(ctx, deployment); err != nil {
			return ctrl.Result{}, err
		}
	} else if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func int32Ptr(i int32) *int32 { return &i }

// SetupWithManager sets up the controller with the Manager.
func (r *TenantReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appv1alpha1.Tenant{}).
		Owns(&apps.Deployment{}).
		Complete(r)
}
