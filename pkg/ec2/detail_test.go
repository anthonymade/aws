package ec2

import (
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"reflect"
	"testing"
	"time"
	"utils/aws/pkg/table"
)

func TestDefaultNoResults(t *testing.T) {
	t.Run("Default, no results", func(t *testing.T) {
		dio := ec2.DescribeInstancesOutput{
			Reservations: []types.Reservation{},
		}
		table, err := Default(&dio, false)
		if err != nil {
			t.Fatalf("error creating table with no results: %v", err)
		}
		expectedHeader := []string{"name", "id", "privateIp", "az", "state", "type", "launched", "imageId"}
		if !reflect.DeepEqual(table.Header, expectedHeader) {
			t.Errorf("header got %+v, want %+v", table.Header, expectedHeader)
		}
		if len(table.Rows) != 0 {
			t.Errorf("number of rows, got %v, want 0", len(table.Rows))
		}
	})
}

func TestDefaultSingleInstance(t *testing.T) {
	var data = []struct {
		name                 *string
		id                   string
		privateIp            *string
		az                   string
		state                types.InstanceState
		instanceType         types.InstanceType
		launched             string
		imageId              string
		expectedPrivateIp    string
		expectedState        string
		expectedInstanceType string
		expectedName         string
		withTags             bool
		extraTags            []types.Tag
		expectedTags         []table.Tag
	}{
		{name: mkStrRef("my-instance-with-all-present"), id: "i-123", privateIp: mkStrRef("1.2.3.4"), az: "us-east-1", state: types.InstanceState{Name: types.InstanceStateNameRunning},
			instanceType: types.InstanceTypeT3Micro, launched: "2021-09-26T19:21:42", imageId: "ami-453823", expectedPrivateIp: "1.2.3.4",
			expectedState: "running", expectedInstanceType: "t3.micro", expectedName: "my-instance-with-all-present",
			extraTags:    []types.Tag{{Key: mkStrRef("team"), Value: mkStrRef("blue")}},
			expectedTags: []table.Tag{}},
		{name: mkStrRef("my-instance-with-tags"), id: "i-123", privateIp: mkStrRef("1.2.3.4"), az: "us-east-1", state: types.InstanceState{Name: types.InstanceStateNameRunning},
			instanceType: types.InstanceTypeT3Micro, launched: "2021-09-26T19:21:42", imageId: "ami-453823", expectedPrivateIp: "1.2.3.4",
			expectedState: "running", expectedInstanceType: "t3.micro", expectedName: "my-instance-with-tags",
			extraTags: []types.Tag{{Key: mkStrRef("team"), Value: mkStrRef("blue")}}, withTags: true,
			expectedTags: []table.Tag{{Key: "Name", Value: "my-instance-with-tags"}, {Key: "team", Value: "blue"}}},
		{name: nil, id: "i-9877", privateIp: nil, az: "eu-west-2", state: types.InstanceState{Name: types.InstanceStateNameStopped},
			instanceType: types.InstanceTypeM3Medium, launched: "2021-09-26T19:21:42", imageId: "ami-111111", expectedPrivateIp: "-",
			expectedState: "stopped", expectedInstanceType: "m3.medium", expectedName: "-", extraTags: []types.Tag{}, expectedTags: []table.Tag{}},
	}
	for i, d := range data {
		t.Run(fmt.Sprintf("Default, single instance %v", i), func(t *testing.T) {
			lTime, err := time.Parse(time.RFC3339, d.launched+"Z")
			if err != nil {
				t.Fatalf("error parsing test time: %v", err)
			}
			dio := ec2.DescribeInstancesOutput{Reservations: []types.Reservation{{Instances: []types.Instance{
				createInstance(d.name, d.id, d.privateIp, d.az, d.state, d.instanceType, lTime, d.imageId, d.extraTags),
			}}}}
			table, err := Default(&dio, d.withTags)
			if err != nil {
				t.Fatalf("error creating table with no results: %v", err)
			}
			expectedRow := []string{d.expectedName, d.id, d.expectedPrivateIp, d.az, d.expectedState, d.expectedInstanceType, d.launched, d.imageId}
			if len(table.Rows) != 1 {
				t.Errorf("number of rows, got %v, want 1", len(table.Rows))
			}
			if !reflect.DeepEqual(table.Rows[0], expectedRow) {
				t.Errorf("row got %#v, want %#v", table.Rows[0], expectedRow)
			}
			if !reflect.DeepEqual(table.Tags[0], d.expectedTags) {
				t.Errorf("tags got %#v, want %#v", table.Tags[0], d.expectedTags)
			}
		})
	}
}

func mkStrRef(s string) *string {
	return &s
}

func createInstance(name *string, id string, privateIp *string, az string, state types.InstanceState, instanceType types.InstanceType, launched time.Time, imageId string, tags []types.Tag) types.Instance {
	if name != nil {
		nameTag := types.Tag{
			Key:   mkStrRef("Name"),
			Value: name,
		}
		tags = append(tags, nameTag)
	}
	return types.Instance{
		InstanceId:        &id,
		PrivateIpAddress:  privateIp,
		NetworkInterfaces: []types.InstanceNetworkInterface{},
		Placement:         &types.Placement{AvailabilityZone: &az},
		State:             &state,
		InstanceType:      instanceType,
		LaunchTime:        &launched,
		ImageId:           &imageId,
		Tags:              tags,
	}
}
