package ec2

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func TestFindAmiIdArgs(t *testing.T) {
	var data = []struct {
		testName       string
		search         []string
		expectedResult []string
	}{
		{"no amiIds", []string{"a_name", "more_name", "i-instance"}, []string{}},
		{"an amiId", []string{"ami-123245"}, []string{"ami-123245"}},
		{"2 amiIds", []string{"ami-123245", "something_else*", "ami-7656723"},
			[]string{"ami-123245", "ami-7656723"}},
	}
	for _, d := range data {
		t.Run(d.testName, func(t *testing.T) {
			result := FindAmiIDArgs(d.search)
			if !reflect.DeepEqual(result, d.expectedResult) {
				t.Errorf("got %+v, want %+v", result, d.expectedResult)
			}
		})
	}
}

func TestFindInstanceIdArgs(t *testing.T) {
	var data = []struct {
		testName       string
		search         []string
		expectedResult []string
	}{
		{"no namePatterns", []string{"a_name", "more_name", "ami-instance"}, []string{}},
		{"an instance", []string{"i-123245"}, []string{"i-123245"}},
		{"2 amiIds", []string{"i-123245", "something_else*", "i-7656723"},
			[]string{"i-123245", "i-7656723"}},
	}
	for _, d := range data {
		t.Run(d.testName, func(t *testing.T) {
			result := FindInstanceIDArgs(d.search)
			if !reflect.DeepEqual(result, d.expectedResult) {
				t.Errorf("got %+v, want %+v", result, d.expectedResult)
			}
		})
	}
}

func TestFindNameSearchArgs(t *testing.T) {
	var data = []struct {
		testName       string
		search         []string
		expectedResult []string
	}{
		{"no names", []string{"i-12345435", "ami-instance"}, []string{}},
		{"a name", []string{"instance_name"}, []string{"instance_name"}},
		{"some names", []string{"i-123245", "something_else*", "a_name", "*mongo*"},
			[]string{"something_else*", "a_name", "*mongo*"}},
	}
	for _, d := range data {
		t.Run(d.testName, func(t *testing.T) {
			result := FindNameSearchArgs(d.search)
			if !reflect.DeepEqual(result, d.expectedResult) {
				t.Errorf("got %+v, want %+v", result, d.expectedResult)
			}
		})
	}
}

type instanceFinderMock struct {
	expectedInput *ec2.DescribeInstancesInput
	actualInput   *ec2.DescribeInstancesInput
	output        *ec2.DescribeInstancesOutput
}

func (ifm *instanceFinderMock) DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	ifm.actualInput = params
	return ifm.output, nil
}

type ValidationError struct {
	Message string
}

func (ve ValidationError) Error() string {
	return ve.Message
}

func prettyFilters(filters []types.Filter) string {
	filterStrings := make([]string, len(filters))
	for i, filter := range filters {
		filterStrings[i] = fmt.Sprintf("{%v: %v}", *filter.Name, strings.Join(filter.Values, ", "))
	}
	return fmt.Sprintf("[%v]", strings.Join(filterStrings, ", "))
}

func (ifm instanceFinderMock) validate() error {
	if ifm.actualInput == nil {
		return ValidationError{Message: fmt.Sprintf("DescribeInstances not called or called with nil input, expected %+v", ifm.expectedInput)}
	}
	if !reflect.DeepEqual(ifm.expectedInput.InstanceIds, ifm.actualInput.InstanceIds) {
		return ValidationError{Message: fmt.Sprintf("InstanceIds, expected %v, got %v", ifm.expectedInput.InstanceIds, ifm.actualInput.InstanceIds)}
	}
	if !reflect.DeepEqual(ifm.expectedInput.Filters, ifm.actualInput.Filters) {
		return ValidationError{Message: fmt.Sprintf("Filters, expected %v, got %v", prettyFilters(ifm.expectedInput.Filters), prettyFilters(ifm.actualInput.Filters))}
	}
	return nil
}

func TestGetInstancesById(t *testing.T) {
	var data = []struct {
		instanceIds []string
	}{
		{instanceIds: []string{}},
		{instanceIds: []string{"i-123"}},
		{instanceIds: []string{"i-123", "i-567", "i-9866"}},
	}
	for _, d := range data {
		t.Run(strings.Join(d.instanceIds, " "), func(t *testing.T) {
			output := ec2.DescribeInstancesOutput{}
			expectedInput := ec2.DescribeInstancesInput{InstanceIds: d.instanceIds, Filters: []types.Filter{}}
			mockInstanceFinder := instanceFinderMock{expectedInput: &expectedInput, output: &output}

			result := getInstances(nil, &mockInstanceFinder, d.instanceIds)

			if result != &output {
				t.Error("expected result to point to output returned from mock")
			}
			err := mockInstanceFinder.validate()
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func TestGetInstancesByImageId(t *testing.T) {
	var data = []struct {
		imageIds []string
	}{
		{imageIds: []string{}},
		{imageIds: []string{"ami-123"}},
		{imageIds: []string{"ami-123", "ami-567", "ami-9866"}},
	}
	for _, d := range data {
		t.Run(strings.Join(d.imageIds, " "), func(t *testing.T) {
			output := ec2.DescribeInstancesOutput{}
			filterName := "image-id"
			filters := []types.Filter{}
			if len(d.imageIds) > 0 {
				filters = []types.Filter{{Name: &filterName, Values: d.imageIds}}
			}
			expectedInput := ec2.DescribeInstancesInput{InstanceIds: []string{}, Filters: filters}
			mockInstanceFinder := instanceFinderMock{expectedInput: &expectedInput, output: &output}

			result := getInstances(nil, &mockInstanceFinder, d.imageIds)

			if result != &output {
				t.Error("expected result to point to output returned from mock")
			}
			err := mockInstanceFinder.validate()
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func TestGetInstancesByName(t *testing.T) {
	var data = []struct {
		namePatterns []string
	}{
		{namePatterns: []string{}},
		{namePatterns: []string{"olap_database"}},
		{namePatterns: []string{"proxy", "web_server_*", "*database*"}},
	}
	for _, d := range data {
		t.Run(strings.Join(d.namePatterns, " "), func(t *testing.T) {
			output := ec2.DescribeInstancesOutput{}
			filterName := "tag:Name"
			filters := []types.Filter{}
			if len(d.namePatterns) > 0 {
				filters = []types.Filter{{Name: &filterName, Values: d.namePatterns}}
			}
			expectedInput := ec2.DescribeInstancesInput{InstanceIds: []string{}, Filters: filters}
			mockInstanceFinder := instanceFinderMock{expectedInput: &expectedInput, output: &output}

			result := getInstances(nil, &mockInstanceFinder, d.namePatterns)

			if result != &output {
				t.Error("expected result to point to output returned from mock")
			}
			err := mockInstanceFinder.validate()
			if err != nil {
				t.Error(err)
			}
		})
	}
}
