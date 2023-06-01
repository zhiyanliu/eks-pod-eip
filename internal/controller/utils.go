package controller

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}

func getAwsEniId(awsSession *session.Session, privateIP string) (string, error) {
	ec2Svc := ec2.New(awsSession)

	result, err := ec2Svc.DescribeNetworkInterfaces(&ec2.DescribeNetworkInterfacesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("addresses.private-ip-address"),
				Values: []*string{aws.String(privateIP)},
			},
		},
	})

	if err != nil {
		return "", err
	}

	if len(result.NetworkInterfaces) == 0 {
		return "", nil
	}

	return *result.NetworkInterfaces[0].NetworkInterfaceId, nil
}
