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
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/zhiyanliu/eks-pod-eip/internal"
	"github.com/zhiyanliu/eks-pod-eip/internal/ipam"
)

const (
	finalizerName = "rp.amazonaws.com/eks-pod-eip-assign"
)

type EKSPodEipAssignReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	IPAM   *ipam.IPAddressManager
}

//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *EKSPodEipAssignReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	logger.V(1).Info(fmt.Sprintf("----------- pod event received: %v\n", req))

	var pod corev1.Pod
	if err := r.Get(ctx, req.NamespacedName, &pod); err != nil {
		if apierrors.IsNotFound(err) {
			// ignore not-found error, the pod has been deleted
			return ctrl.Result{}, nil
		}
		logger.V(1).Error(err, fmt.Sprintf("unable to fetch Pod %s: %v", req.NamespacedName, err))
		return ctrl.Result{}, err
	}

	var ns corev1.Namespace
	if err := r.Get(ctx, types.NamespacedName{Name: pod.GetNamespace()}, &ns); err != nil {
		if apierrors.IsNotFound(err) {
			// doesn't make sense, but anyway
			return ctrl.Result{}, nil
		}
		logger.V(1).Error(err, fmt.Sprintf("unable to fetch Namespace %s: %v", pod.GetNamespace(), err))
		return ctrl.Result{}, err
	}

	enabled := ns.GetLabels()[internal.NamespacePodEipAllocationEnabledLabel]
	if enabled == "true" && pod.DeletionTimestamp.IsZero() {
		if pod.Status.PodIP == "" {
			// pod is not ready yet, wait the ip address is allocated to the pod
			return ctrl.Result{}, nil
		}

		// append the finalizer to the pod if not exist
		if !containsString(pod.Finalizers, finalizerName) {
			pod.Finalizers = append(pod.Finalizers, finalizerName)
			if err := r.Update(ctx, &pod); err != nil {
				return ctrl.Result{}, err
			}
		}

		if awsEIPAllocationId, err := r.IPAM.EnsureAllocation(&pod); err != nil {
			return ctrl.Result{}, err
		} else {
			logger.V(1).Info(fmt.Sprintf(
				"pod %s is assigned with the aws EIP allocation %s", req.NamespacedName, awsEIPAllocationId))
		}
	} else { // pod EIP allocation is disabled or the pod is being deleted
		if awsEIPAllocationId, err := r.IPAM.ReleaseAllocation(&pod); err != nil {
			return ctrl.Result{}, err
		} else if awsEIPAllocationId != "" {
			logger.V(1).Info(fmt.Sprintf(
				"pod %s is unassigned from the aws EIP allocation %s", req.NamespacedName, awsEIPAllocationId))
		}

		// remove the finalizer from the pod
		pod.Finalizers = removeString(pod.Finalizers, finalizerName)
		if err := r.Update(ctx, &pod); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *EKSPodEipAssignReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("eks-pod-eip-assign-controller").
		Watches(
			&source.Kind{Type: &corev1.Pod{}},
			&handler.EnqueueRequestForObject{},
			builder.WithPredicates(EipAssignPodPredicate{})).
		Watches(
			&source.Kind{Type: &corev1.Namespace{}},
			handler.EnqueueRequestsFromMapFunc(r.EipAssignNamespaceMapFunc),
			builder.WithPredicates(EipAssignNamespacePredicate{})).
		Complete(r)
}

func (r *EKSPodEipAssignReconciler) EipAssignNamespaceMapFunc(obj client.Object) []ctrl.Request {
	ns, ok := obj.(*corev1.Namespace)
	if !ok {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*2)
	defer cancel()

	logger := log.FromContext(ctx)

	podList := &corev1.PodList{}
	err := r.List(ctx, podList, client.InNamespace(ns.Name))
	if err != nil {
		logger.V(1).Error(err, fmt.Sprintf(
			"could not list pod in namespace %s: %v. change to Namespace %s will not be reconciled.",
			ns.Name, err, ns.Name))
		return nil
	}

	requests := make([]reconcile.Request, len(podList.Items))

	for idx, pod := range podList.Items {
		requests[idx] = reconcile.Request{NamespacedName: types.NamespacedName{
			Namespace: pod.Namespace,
			Name:      pod.Name,
		}}
	}

	return requests
}
