package ec2

import (
	"sort"
	"strings"
	"utils/aws/pkg/table"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func Default(ec2Output *ec2.DescribeInstancesOutput, withTags bool) (*table.FixedWidthFont, error) {
	var instances = table.New([]string{"name", "id", "privateIp", "az", "state", "type", "launched", "imageId"})
	for _, reservation := range ec2Output.Reservations {
		for _, instance := range reservation.Instances {
			launchTime := *instance.LaunchTime
			var privateIP = "-"
			if instance.PrivateIpAddress != nil {
				privateIP = *instance.PrivateIpAddress
			}
			nameTag := tagValueByKey(instance.Tags, "Name")
			tags := []table.Tag{}
			if withTags {
				tags = tableTags(instance.Tags)
			}
			var name = "-"
			if nameTag != nil {
				name = *nameTag
			}
			err := instances.AddRow([]string{
				name,
				*instance.InstanceId,
				privateIP,
				//*instance.NetworkInterfaces,
				*instance.Placement.AvailabilityZone,
				string(instance.State.Name),
				string(instance.InstanceType),
				launchTime.Format("2006-01-02T15:04:05"),
				*instance.ImageId,
			}, tags)
			if err != nil {
				return nil, err
			}
		}
	}
	return &instances, nil
}

func tagValueByKey(tags []types.Tag, key string) *string {
	for _, tag := range tags {
		if *tag.Key == key {
			return tag.Value
		}
	}
	return nil
}

func tableTags(tags []types.Tag) []table.Tag {
	sort.Slice(tags, func(i int, j int) bool {
		return strings.Compare(*tags[i].Key, *tags[j].Key) == -1
	})
	tableTags := make([]table.Tag, 0, 5)
	for _, tag := range tags {
		tableTags = append(tableTags, table.Tag{Key: *tag.Key, Value: *tag.Value})
	}
	return tableTags
}
