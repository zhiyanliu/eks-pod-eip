package ipam

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/zhiyanliu/eks-pod-eip/internal"
)

type IPAddressManager struct {
	k8sClient client.Client
}

func NewIPAddressManager(k8sClient client.Client) *IPAddressManager {
	return &IPAddressManager{
		k8sClient: k8sClient,
	}
}

func (m *IPAddressManager) EnsureAllocation(pod *corev1.Pod) (string, error) {
	if pod == nil {
		return "", fmt.Errorf("pod is nil")
	}

	fmt.Sprintf("%s/%s", pod.GetNamespace(), pod.GetName())

	eipAllocationId

	if preferredEIPAllocationId, exists := pod.GetAnnotations()[internal.PodEipAllocationIdAnnotation]; exists {
		preferredEIPAllocationId
		m.associate(pod)
	}

	return "", nil
}

func (m *IPAddressManager) ReleaseAllocation(pod *corev1.Pod) (string, error) {
	return "", nil
}

func (m *IPAddressManager) associate(pod *corev1.Pod, eipAllocationId string) (string, error) {
	return "", nil
}
