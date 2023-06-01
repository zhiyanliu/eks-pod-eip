package controller

import (
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/zhiyanliu/eks-pod-eip/internal"
)

var _ predicate.Predicate = &EipAssignPodPredicate{}

type EipAssignPodPredicate struct {
	predicate.Funcs
}

func (p EipAssignPodPredicate) Create(e event.CreateEvent) bool {
	if e.Object.GetNamespace() == "kube-system" {
		return false
	}

	_, ok := e.Object.(*corev1.Pod)
	if !ok {
		return false
	}

	return true
}

func (p EipAssignPodPredicate) Update(e event.UpdateEvent) bool {
	oldPod, ok := e.ObjectOld.(*corev1.Pod)
	if !ok {
		return false
	}
	newPod, ok := e.ObjectNew.(*corev1.Pod)
	if !ok {
		return false
	}

	if newPod.Namespace == "kube-system" {
		return false
	}

	if oldPod.Status.PodIP != newPod.Status.PodIP {
		return true
	}

	if !newPod.DeletionTimestamp.IsZero() {
		return true
	}

	oldEipAllocationIdValue, _ := e.ObjectOld.GetAnnotations()[internal.PodEipAllocationIdAnnotation]
	newEipAllocationIdValue, _ := e.ObjectNew.GetAnnotations()[internal.PodEipAllocationIdAnnotation]

	if oldEipAllocationIdValue != newEipAllocationIdValue {
		return true
	}

	return false
}

//func (p EipAssignPodPredicate) Delete(e event.DeleteEvent) bool {
//	if e.Object.GetNamespace() == "kube-system" {
//		return false
//	}
//
//	_, ok := e.Object.(*corev1.Pod)
//	if !ok {
//		return false
//	}
//
//	return true
//}

func (p EipAssignPodPredicate) Generic(e event.GenericEvent) bool {
	if e.Object.GetNamespace() == "kube-system" {
		return false
	}

	_, ok := e.Object.(*corev1.Pod)
	if !ok {
		return false
	}

	return true
}

var _ predicate.Predicate = &EipAssignNamespacePredicate{}

type EipAssignNamespacePredicate struct {
	predicate.Funcs
}

func (p EipAssignNamespacePredicate) Create(e event.CreateEvent) bool {
	return false
}

func (p EipAssignNamespacePredicate) Update(e event.UpdateEvent) bool {
	_, ok := e.ObjectOld.(*corev1.Namespace)
	if !ok {
		return false
	}
	newNs, ok := e.ObjectNew.(*corev1.Namespace)
	if !ok {
		return false
	}

	if newNs.GetNamespace() == "kube-system" {
		return false
	}

	oldValue, _ := e.ObjectOld.GetLabels()[internal.NamespacePodEipAllocationEnabledLabel]
	newValue, _ := e.ObjectNew.GetLabels()[internal.NamespacePodEipAllocationEnabledLabel]

	return oldValue != newValue
}

func (p EipAssignNamespacePredicate) Delete(e event.DeleteEvent) bool {
	return false
}

func (p EipAssignNamespacePredicate) Generic(e event.GenericEvent) bool {
	if e.Object.GetNamespace() == "kube-system" {
		return false
	}

	_, ok := e.Object.(*corev1.Namespace)
	if !ok {
		return false
	}

	return true
}
