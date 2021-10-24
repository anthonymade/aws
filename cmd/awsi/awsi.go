package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"utils/aws/pkg/ec2"

	"github.com/aws/aws-sdk-go-v2/config"
)

type arguments struct {
	noHeadings bool
	tags       bool
	less       bool
	search     []string
}

func parseFlags(cmdName string, args []string) (arguments, string, error) {
	var a arguments
	var buf bytes.Buffer
	flags := flag.NewFlagSet(cmdName, flag.ContinueOnError)
	flags.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s: [OPTIONS...] [name-tag-expression...] [instance-id...] [ami-id...]\n\n", os.Args[0])
		fmt.Fprint(flag.CommandLine.Output(), "Find aws ec2 instances in an account.\nUse aws-vault or equivalent to provide credentials and select the account.\n\n")
		flags.PrintDefaults()
	}
	flags.SetOutput(&buf)
	flags.BoolVar(&a.noHeadings, "n", false, "")
	flags.BoolVar(&a.noHeadings, "no-header", false, "do not output header")
	flags.BoolVar(&a.tags, "t", false, "")
	flags.BoolVar(&a.tags, "tags", false, "print tags")
	err := flags.Parse(args)
	if err != nil {
		return a, buf.String(), err
	}
	a.search = flags.Args()
	return a, buf.String(), nil
}

func main() {
	noTimestamp := 0
	stderr := log.New(os.Stderr, "", noTimestamp)
	args, output, err := parseFlags(os.Args[0], os.Args[1:])
	if err != nil && errors.Is(err, flag.ErrHelp) {
		println(output)
		os.Exit(0)
	}
	if err != nil {
		stderr.Fatal(err)
	}
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		stderr.Fatal(err)
	}
	instances := ec2.GetInstances(ctx, cfg, args.search)
	table, err := ec2.Default(instances, args.tags)
	if err != nil {
		stderr.Fatal(err)
	}
	table.Print(os.Stdout, !args.noHeadings, args.tags)
}
