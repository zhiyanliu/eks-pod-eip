package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func getAwsSession() *session.Session {
	return session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
}

func getEksVpcId(awsSession *session.Session) string {
	if awsSession == nil {
		panic("aws session is nil")
	}

	ec2metadataSvc := ec2metadata.New(awsSession)

	doc, err := ec2metadataSvc.GetInstanceIdentityDocument()
	if err != nil {
		panic(fmt.Errorf("failed to get instance identity document, "+
			"the program needs to running on EC2 or Fargate in an aws vpc: %v", err))
	}

	instanceID := doc.InstanceID

	ec2Svc := ec2.New(awsSession)

	result, err := ec2Svc.DescribeInstances(&ec2.DescribeInstancesInput{
		InstanceIds: []*string{
			aws.String(instanceID),
		},
	})
	if err != nil {
		panic(err)
	}

	for _, res := range result.Reservations {
		for _, instance := range res.Instances {
			return *instance.VpcId
		}
	}

	panic(fmt.Errorf("no vpc id found for instance %s", instanceID))
}
