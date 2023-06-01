package ipam

type IPAddressStore interface {
	AssociateEIPAllocationId(podNamespace, podName, podIP, eipAllocationId string) (string, error)
	ReleaseEIPAllocationId(podNamespace, podName, podIP, eipAllocationId string) (string, error)

	GetAssociatedEIPAllocationId(podNamespace, podName, podIP, eipAllocationId string) (string, error)
	GetAllAssociatedEIPAllocationIds() ([]string, error)
	GetAvailableEIPAllocationIds() ([]string, error)
}
