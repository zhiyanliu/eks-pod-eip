package ipam

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	corev1 "k8s.io/api/core/v1"

	"github.com/zhiyanliu/eks-pod-eip/internal"
)

type IPAddressManager struct {
	awsSession *session.Session
	vpcId      string
}

func NewIPAddressManager(awsSession *session.Session) *IPAddressManager {
	if awsSession == nil {
		panic("aws session is nil")
	}

	return &IPAddressManager{
		awsSession: awsSession,
	}
}

func (m *IPAddressManager) AllocateEip(pod *corev1.Pod) (string, error) {
	if preferredEIPAllocationId, exists := pod.GetAnnotations()[internal.PodEipAllocationIdAnnotation]; exists {
		return preferredEIPAllocationId, nil
	}
	return m.createAwsEip()
}

func (m *IPAddressManager) ReleaseEip(pod *corev1.Pod) (string, error) {
	return "", nil
}

func (m *IPAddressManager) createAwsEip() (string, error) {
	ec2Svc := ec2.New(m.awsSession)

	eipAllocation, err := ec2Svc.AllocateAddress(&ec2.AllocateAddressInput{
		Domain: aws.String("vpc"),
	})
	if err != nil {
		return "", err
	}

	return *eipAllocation.AllocationId, nil
}
