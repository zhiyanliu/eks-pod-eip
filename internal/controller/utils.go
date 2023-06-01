package controller

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

//
//func queryEniIDbyPrivateIP(privateIP string) (eniID string, err error) {
//	awsSession := session.Must(session.NewSessionWithOptions(session.Options{
//		SharedConfigState: session.SharedConfigEnable,
//	}))
//
//	svc := ec2.New(awsSession)
//
//	result, err := svc.DescribeNetworkInterfaces(&ec2.DescribeNetworkInterfacesInput{
//		Filters: []*ec2.Filter{
//			{
//				Name:   aws.String("addresses.private-ip-address"),
//				Values: []*string{aws.String(privateIP)},
//			},
//		},
//	})
//
//	if err != nil {
//		log.Println("Error describing network interfaces:", err)
//		return nil
//	}
//
//	if len(result.NetworkInterfaces) == 0 {
//		log.Println("No ENI found associated with the private IP:", privateIp)
//		return nil
//	}
//
//	return result.NetworkInterfaces[0].NetworkInterfaceId
//}
