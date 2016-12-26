package ec2util

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/aws"
	"fmt"
	"log"
)

type Ec2Service interface{
	InstancesByRegionTagAndValue (region string, tag string, value string) map[string]*ec2.Instance
}

type Ec2ServiceImpl struct {
	Session *session.Session
	Region string
}


func (self *Ec2ServiceImpl) InstancesByRegionTagAndValue (region string, tag string, value string) map[string]*ec2.Instance{
	//nameFilter := "*" // os.Args[1]
	params := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("tag:Lane"),
				Values: []*string{
					aws.String("*"),
					//aws.String(strings.Join([]string{"*", nameFilter, "*"}, "")),
				},
			},
		},
	}

	//
	svc := ec2.New(self.Session, &aws.Config{Region: aws.String(self.Region)})

	resp, err := svc.DescribeInstances(params)
	if err != nil {
		fmt.Println("there was an error listing instances in", self.Region, err.Error())
		log.Fatal(err.Error())
	}

	instances := make(map[string] *ec2.Instance, 100);

	for i :=0; i < len(resp.Reservations); i++ {
		res := resp.Reservations[i]
		for _, instance := range res.Instances {
			tags := instance.Tags;
			for _, tag := range tags {
				if *tag.Key == "Name" {
					instances[*tag.Value] = instance
				}
			}

		}
	}

	return instances
}