package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseArgs(t *testing.T) {
	var data = []struct {
		args []string
		opts arguments
		err  string
	}{
		{[]string{"-no-header", "i-1234"},
			arguments{noHeadings: true, tags: false, search: []string{"i-1234"}}, ""},
		{[]string{"-no-header", "i-1234", "i-5678"},
			arguments{noHeadings: true, search: []string{"i-1234", "i-5678"}}, ""},
		{[]string{"-n", "i-1234", "i-5678"},
			arguments{noHeadings: true, search: []string{"i-1234", "i-5678"}}, ""},
		{[]string{"instance_name"},
			arguments{noHeadings: false, search: []string{"instance_name"}}, ""},
		{[]string{"-unknown"},
			arguments{noHeadings: false, search: []string{}}, "not defined: -unknown"},
		{[]string{"-t", "name"},
			arguments{tags: true, search: []string{"name"}}, ""},
		{[]string{"--tags", "name"},
			arguments{tags: true, search: []string{"name"}}, ""},
	}
	for _, d := range data {
		t.Run(strings.Join(d.args, " "), func(t *testing.T) {
			a, output, err := parseFlags("prog", d.args)
			if err != nil && d.err == "" {
				t.Fatalf("err got %v, want nil", err)
			}
			if err == nil && d.err != "" {
				t.Fatalf("expected error, did not get one, %q", d.err)
			}
			if output != "" && d.err == "" {
				t.Fatalf("output got %q, want empty", output)
			}
			if d.err != "" && !strings.Contains(output, d.err) {
				t.Fatalf("expected output to contain %q, output was %q", d.err, output)
			}
			if d.err == "" && !reflect.DeepEqual(a, d.opts) {
				t.Fatalf("options got %+v, want %+v", a, d.opts)
			}
		})
	}
}
