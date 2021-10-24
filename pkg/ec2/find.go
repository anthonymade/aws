package ec2

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func GetInstances(ctx context.Context, cfg aws.Config, search []string) *ec2.DescribeInstancesOutput {
	return getInstances(ctx, ec2.NewFromConfig(cfg), search)
}

type instanceFinder interface {
	DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
}

func getInstances(ctx context.Context, finder instanceFinder, search []string) *ec2.DescribeInstancesOutput {
	filters := make([]types.Filter, 0, 2)
	names := FindNameSearchArgs(search)
	if len(names) > 0 {
		filters = append(filters, filter("tag:Name", names))
	}
	amis := FindAmiIDArgs(search)
	if len(amis) > 0 {
		filters = append(filters, filter("image-id", amis))
	}
	input := ec2.DescribeInstancesInput{InstanceIds: FindInstanceIDArgs(search), Filters: filters}
	output, err := finder.DescribeInstances(ctx, &input)
	if err != nil {
		noTimestamp := 0
		stderr := log.New(os.Stderr, "", noTimestamp)
		stderr.Fatal(err)
	}
	return output
}

func filter(name string, values []string) types.Filter {
	return types.Filter{Name: &name, Values: values}
}

func findAll(search []string, predicate func(string) bool) []string {
	var result = make([]string, 0, len(search))
	for _, arg := range search {
		if predicate(arg) {
			result = append(result, arg)
		}
	}
	return result
}

func FindAmiIDArgs(search []string) []string {
	return findAll(search, func(s string) bool {
		return strings.HasPrefix(s, "ami-")
	})
}

func FindInstanceIDArgs(search []string) []string {
	return findAll(search, func(s string) bool {
		return strings.HasPrefix(s, "i-")
	})
}

func FindNameSearchArgs(search []string) []string {
	return findAll(search, func(s string) bool {
		return !(strings.HasPrefix(s, "i-") || strings.HasPrefix(s, "ami-"))
	})
}
