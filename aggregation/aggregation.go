package aggregation

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/sts"
	"strings"
)

func createEC2Service(profile string, assumeRole string, region string) *ec2.EC2 {
	var sess *session.Session
	var svc *ec2.EC2

	// Create Session
	if profile == "" && assumeRole == "" { // Profile and AssumeRole are not specified
		sess = session.Must(session.NewSession())
		svc = ec2.New(sess, aws.NewConfig().WithRegion(region))
	} else if profile != "" && assumeRole == "" { // Only Profile is specified, but no AssumeRole
		sess = session.Must(session.NewSessionWithOptions(session.Options{Profile: profile}))
		svc = ec2.New(sess, aws.NewConfig().WithRegion(region))
	} else if profile == "" && assumeRole != "" { // Only AssumeRole is specified, but no Profile
		sess = session.Must(session.NewSession())
		var assumeRoler *sts.STS = sts.New(sess)
		var creds *credentials.Credentials = stscreds.NewCredentialsWithClient(assumeRoler, assumeRole)
		svc = ec2.New(sess, aws.NewConfig().WithRegion(region).WithCredentials(creds))
	} else if profile != "" && assumeRole != "" { // Both Profile and AssumeRole exist
		sess = session.Must(session.NewSessionWithOptions(session.Options{Profile: profile}))
		var assumeRoler *sts.STS = sts.New(sess)
		var creds *credentials.Credentials = stscreds.NewCredentialsWithClient(assumeRoler, assumeRole)
		svc = ec2.New(sess, aws.NewConfig().WithRegion(region).WithCredentials(creds))
	}

	return svc
}

func getAllRegeons(svc *ec2.EC2) (*[]string, error) {
	var out *ec2.DescribeRegionsOutput
	var err error

	out, err = svc.DescribeRegions(&ec2.DescribeRegionsInput{DryRun: aws.Bool(false)})
	if err != nil {
		return nil, err
	}

	var regions []string = make([]string, 0)
	var region *ec2.Region
	var regionName *string
	for _, region = range out.Regions {
		regionName = region.RegionName
		regions = append(regions, aws.StringValue(regionName))
	}

	return &regions, nil
}

func describeInstances(svc *ec2.EC2) (*[][]string, error) {
	var out *ec2.DescribeInstancesOutput
	var err error
	var input *ec2.DescribeInstancesInput

	input = &ec2.DescribeInstancesInput{
		DryRun:     aws.Bool(false),
		MaxResults: aws.Int64(1000),
	}

	out, err = svc.DescribeInstances(input)
	if err != nil {
		return nil, err
	}

	var reservations []*ec2.Reservation = out.Reservations
	var reservation *ec2.Reservation
	var instances []*ec2.Instance
	var instance *ec2.Instance

	var result [][]string = [][]string{}
	var row []string
	var placement *ec2.Placement
	for _, reservation = range reservations {
		instances = reservation.Instances

		for _, instance = range instances {
			row = []string{}
			placement = instance.Placement
			row = append(row, aws.StringValue(placement.AvailabilityZone)) // The 1st col: AvailabilityZone
			row = append(row, aws.StringValue(instance.InstanceType))      // The 2nd col: InstanceType
			row = append(row, aws.StringValue(instance.Platform))          // The 3rd col: Platform
			row = append(row, aws.StringValue(placement.Tenancy))          // The 4th col: Tenancy
			result = append(result, row)
		}
	}

	for out.NextToken != nil {
		input.NextToken = out.NextToken
		out, err = svc.DescribeInstances(input)
		if err != nil {
			return nil, err
		}

		reservations = out.Reservations
		for _, reservation = range reservations {
			instances = reservation.Instances

			for _, instance = range instances {
				row = []string{}
				placement = instance.Placement
				row = append(row, aws.StringValue(placement.AvailabilityZone)) // The 1st col: AvailavilityZone
				row = append(row, aws.StringValue(instance.InstanceType))      // The 2nd col: InstanceType
				row = append(row, aws.StringValue(instance.Platform))          // The 3rd col: Platform
				row = append(row, aws.StringValue(placement.Tenancy))          // The 4th col: Tenancy
				result = append(result, row)
			}
		}
	}

	return &result, nil

}

func Regions(profile string, assumeRole string) (*[]string, error) {
	var svc *ec2.EC2 = createEC2Service(profile, assumeRole, "us-east-1")
	var regions *[]string
	var err error
	regions, err = getAllRegeons(svc)
	if err != nil {
		return nil, err
	}
	return regions, nil
}

func Instances(profile string, assumeRole string, regions *[]string) (*[][]string, error) {
	var region string
	var svc *ec2.EC2
	var instances *[][]string
	var err error
	var result [][]string = [][]string{}

	for _, region = range *regions {
		svc = createEC2Service(profile, assumeRole, region)
		instances, err = describeInstances(svc)
		if err != nil {
			return nil, err
		}
		result = append(result, *instances...)
	}
	return &result, nil
}

func DoAggregate(instances *[][]string) *map[string]int {
	var result map[string]int = map[string]int{}
	var instance []string
	var key string

	for _, instance = range *instances {
		key = strings.Join(instance, ":")
		result[key] += 1
	}

	return &result
}
