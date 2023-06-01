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

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	ekspodeipv1 "github.com/zhiyanliu/eks-pod-eip/api/v1"
	"github.com/zhiyanliu/eks-pod-eip/internal"
	"github.com/zhiyanliu/eks-pod-eip/internal/ipam"
)

const (
	finalizerName = "rp.amazonaws.com/eks-pod-eip-assign"
)

type EksPodEipAssignReconciler struct {
	client.Client
	Scheme               *runtime.Scheme
	IPAM                 *ipam.IPAddressManager
	AssociationNamespace string
	VpcId                string
}

//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=core,resources=pods/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch

//+kubebuilder:rbac:groups=ekspodeip.rp.amazonaws.com,resources=ekspodeipassociations,verbs=get;update;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *EksPodEipAssignReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
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

		if eipAllocationID, err := r.ensureAssociation(&ctx, &logger, &pod); err != nil {
			logger.V(1).Error(err, fmt.Sprintf(
				"unable to ensure the aws EIP association for pod %s", req.NamespacedName))

			return ctrl.Result{}, err
		} else {
			logger.V(1).Info(fmt.Sprintf(
				"pod %s is assigned with the aws EIP association %s", req.NamespacedName, eipAllocationID))
		}
	} else { // pod EIP allocation is disabled or the pod is being deleted
		if eipAllocationID, err := r.releaseAssociation(&ctx, &logger, &pod); err != nil {
			logger.V(1).Error(err, fmt.Sprintf(
				"unable to release the aws EIP association for pod %s", req.NamespacedName))

			return ctrl.Result{}, err
		} else if eipAllocationID != "" {
			logger.V(1).Info(fmt.Sprintf(
				"pod %s is released from the aws EIP association %s", req.NamespacedName, eipAllocationID))
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
func (r *EksPodEipAssignReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r.IPAM == nil {
		return fmt.Errorf("ipam is not set")
	}
	if r.VpcId == "" {
		panic("vpc id is empty")
	}

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

func (r *EksPodEipAssignReconciler) EipAssignNamespaceMapFunc(obj client.Object) []ctrl.Request {
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

func (r *EksPodEipAssignReconciler) eipAssociationNamespace(pod *corev1.Pod) string {
	associationNamespace := r.AssociationNamespace
	if associationNamespace == "" {
		associationNamespace = pod.GetNamespace()
	}

	return associationNamespace
}

func (r *EksPodEipAssignReconciler) eipAssociationName(pod *corev1.Pod) string {
	return fmt.Sprintf("eip-asso-%s-%s", pod.GetNamespace(), pod.GetName())
}

func (r *EksPodEipAssignReconciler) ensureAssociation(
	ctx *context.Context, logger *logr.Logger, pod *corev1.Pod) (string, error) {

	if pod == nil {
		return "", fmt.Errorf("pod is nil")
	}

	var eipAssociation ekspodeipv1.EksPodEipAssociation
	if err := r.Get(
		*ctx,
		types.NamespacedName{Name: r.eipAssociationName(pod), Namespace: r.eipAssociationNamespace(pod)},
		&eipAssociation); err == nil { // the association resource exists, need to delete it first
		logger.V(1).Info(fmt.Sprintf("EksPodEipAssociation %s/%s already exists, delete it first",
			r.eipAssociationNamespace(pod), r.eipAssociationName(pod)))

		if err = r.deleteAssociation(ctx, logger, &eipAssociation); err != nil {
			logger.V(1).Error(err, fmt.Sprintf("unable to delete EksPodEipAssociation %s/%s",
				eipAssociation.Namespace, eipAssociation.Name))

			return "", err
		}
	} else if !apierrors.IsNotFound(err) { // error happens
		logger.V(1).Error(err, fmt.Sprintf("unable to fetch EksPodEipAssociation %s/%s: %v",
			r.eipAssociationNamespace(pod), r.eipAssociationName(pod), err))

		return "", fmt.Errorf("unable to fetch EksPodEipAssociation %s/%s: %v",
			r.eipAssociationNamespace(pod), r.eipAssociationName(pod), err)
	}

	// create the association resource
	newEipAssociation, err := r.createAssociation(ctx, logger, pod)
	if err != nil {
		logger.V(1).Error(err, fmt.Sprintf("unable to create EksPodEipAssociation %s/%s",
			newEipAssociation.Namespace, newEipAssociation.Name))

		return "", fmt.Errorf("unable to create EksPodEipAssociation %s/%s: %v",
			newEipAssociation.Namespace, newEipAssociation.Name, err)
	}

	return newEipAssociation.Spec.EipAllocationId, nil
}

func (r *EksPodEipAssignReconciler) releaseAssociation(
	ctx *context.Context, logger *logr.Logger, pod *corev1.Pod) (string, error) {

	if pod == nil {
		return "", fmt.Errorf("pod is nil")
	}

	return "", nil
}

func (r *EksPodEipAssignReconciler) createAssociation(
	ctx *context.Context, logger *logr.Logger, pod *corev1.Pod) (*ekspodeipv1.EksPodEipAssociation, error) {

	var eipAssociation ekspodeipv1.EksPodEipAssociation

	eipAssociation.Name = r.eipAssociationName(pod)
	eipAssociation.Namespace = r.eipAssociationNamespace(pod)
	eipAssociation.Spec = ekspodeipv1.EksPodEipAssociationSpec{
		PodNamespace: pod.GetNamespace(),
		PodName:      pod.GetName(),
		PrivateIP:    pod.Status.PodIP,
	}

	// allocate an EIP
	if eipAllocationId, err := r.IPAM.AllocateEip(pod); err != nil {
		return nil, fmt.Errorf("unable to allocate EIP for pod %s/%s: %v",
			pod.GetNamespace(), pod.GetName(), err)
	} else {
		eipAssociation.Spec.EipAllocationId = eipAllocationId
	}

	ownerRef := metav1.OwnerReference{
		APIVersion: pod.APIVersion,
		Kind:       pod.Kind,
		Name:       pod.Name,
		UID:        pod.UID,
		Controller: new(bool), // false
	}
	eipAssociation.OwnerReferences = append(eipAssociation.OwnerReferences, ownerRef)

	logger.V(1).Info(fmt.Sprintf("creating EksPodEipAssociation %s/%s",
		r.eipAssociationNamespace(pod), r.eipAssociationName(pod)))

	if err := r.Create(*ctx, &eipAssociation); err != nil {
		return nil, err
	}

	logger.V(1).Info(fmt.Sprintf("EksPodEipAssociation %s/%s created",
		r.eipAssociationNamespace(pod), r.eipAssociationName(pod)))

	return &eipAssociation, nil
}

func (r *EksPodEipAssignReconciler) updateAssociation(
	ctx *context.Context, logger *logr.Logger, pod *corev1.Pod, eipAllocationId string) (string, error) {
	return "", nil
}

func (r *EksPodEipAssignReconciler) deleteAssociation(
	ctx *context.Context, logger *logr.Logger, eipAssociation *ekspodeipv1.EksPodEipAssociation) error {

	logger.V(1).Info(fmt.Sprintf("deleting EksPodEipAssociation %s/%s",
		eipAssociation.GetNamespace(), eipAssociation.GetName()))

	if err := r.Delete(*ctx, eipAssociation); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("unable to delete EksPodEipAssociation %s/%s: %v",
			eipAssociation.GetNamespace(), eipAssociation.GetName(), err)
	}

	logger.V(1).Info(fmt.Sprintf("EksPodEipAssociation %s/%s deleted",
		eipAssociation.GetNamespace(), eipAssociation.GetName()))

	return nil
}
