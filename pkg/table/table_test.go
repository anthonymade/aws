package table

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	var data = []struct {
		header         []string
		expectedWidths []int
	}{
		{[]string{"a", "b"}, []int{1, 1}},
		{[]string{"123", "456789"}, []int{3, 6}},
	}
	for _, d := range data {
		t.Run(strings.Join(d.header, " "), func(t *testing.T) {
			fwfTable := New(d.header)
			if !reflect.DeepEqual(fwfTable.Header, d.header) {
				t.Errorf("Header got %+v, want %+v", fwfTable.Header, d.header)
			}
			if !reflect.DeepEqual(fwfTable.widths, d.expectedWidths) {
				t.Errorf("widths got %+v, want %+v", fwfTable.widths, d.expectedWidths)
			}
		})
	}
}

func TestAddRow(t *testing.T) {
	var data = []struct {
		header         []string
		rows           [][]string
		expectedWidths []int
	}{
		{[]string{"col1", "col2"}, [][]string{
			{"row1", "row1"},
		}, []int{4, 4}},
		{[]string{"col", "col"}, [][]string{
			{"longer", "moar width here!"},
		}, []int{6, 16}},
		{[]string{"a", "b"}, [][]string{
			{"some text", "other text!"},
			{"short", "x"},
		}, []int{9, 11}},
	}
	for _, d := range data {
		t.Run(strings.Join(d.header, " "), func(t *testing.T) {
			fwfTable := New(d.header)
			for _, r := range d.rows {
				err := fwfTable.AddRow(r, []Tag{})
				if err != nil {
					t.Fatalf("error adding row: %v", err)
				}
			}
			if !reflect.DeepEqual(fwfTable.Header, d.header) {
				t.Errorf("Header got %+v, want %+v", fwfTable.Header, d.header)
			}
			if !reflect.DeepEqual(fwfTable.widths, d.expectedWidths) {
				t.Errorf("widths got %+v, want %+v", fwfTable.widths, d.expectedWidths)
			}
			if !reflect.DeepEqual(fwfTable.Rows, d.rows) {
				t.Errorf("Rows got %+v, want %+v", fwfTable.Rows, d.rows)
			}
			for _, rowTags := range fwfTable.Tags {
				if rowTags == nil || len(rowTags) != 0 {
					t.Errorf("Tags got %+v, want empty slice", rowTags)
				}
			}
			if len(fwfTable.Rows) != len(fwfTable.Tags) {
				t.Errorf("number of Tags %d != number of Rows %d", len(fwfTable.Tags), len(fwfTable.Rows))
			}
		})
	}
}

func TestAddRowErrors(t *testing.T) {
	var data = []struct {
		header      []string
		row         []string
		expectedErr string
	} {
		{[]string{"col1", "col2"}, []string{"row1"}, "bad row: expected 2, got 1"},
		{[]string{"col1"}, []string{"cell1", "cell2"}, "bad row: expected 1, got 2"},
	}
	for _, d := range data {
		t.Run(strings.Join(d.header, " "), func(t *testing.T) {
			fwfTable := New(d.header)
			err := fwfTable.AddRow(d.row, []Tag{})
			if err == nil {
				t.Fatal("expected error")
			}
			if err.Error() != d.expectedErr {
				t.Errorf("err got %+v, want %+v", err, d.expectedErr)
			}
		})
	}
}

func TestAddRowTags(t *testing.T) {
	var data = []struct {
		tags                    [][]Tag
		expectedMaxTagKeyLength int
	} {
		{[][]Tag{
			{},
		}, 0},
		{[][]Tag{
			{Tag{Key: "k", Value: "v"}},
		}, 1},
		{[][]Tag{
			{Tag{Key: "1", Value: "v"}},
			{Tag{Key: "22", Value: "v"}},
		}, 2},
		{[][]Tag{
			{Tag{Key: "longer", Value: "v"}},
			{Tag{Key: "short", Value: "v"}},
		}, 6},
	}
	for _, d := range data {
		t.Run("maxTagKeyLength", func(t *testing.T) {
			fwfTable := New([]string{"one"})
			for _, tags := range d.tags {
				err := fwfTable.AddRow([]string{"row"}, tags)
				if err != nil {
					t.Fatalf("error adding row: %v", err)
				}
			}
			if fwfTable.maxTagKeyLength != d.expectedMaxTagKeyLength {
				t.Errorf("maxTagKeyLength got %d, want %d", fwfTable.maxTagKeyLength, d.expectedMaxTagKeyLength)
			}
			if !reflect.DeepEqual(fwfTable.Tags, d.tags) {
				t.Errorf("Tags got %+v, want %+v", fwfTable.Tags, d.tags)
			}
		})
	}
}

func createTestTable() FixedWidthFont {
	fwfTable := New([]string{"a", "heading2", "3"})
	fwfTable.AddRow([]string{"r1c1", "more", "1"}, []Tag{{Key: "k", Value: "val1"}})
	fwfTable.AddRow([]string{"r2c1", "cellr2", "2"}, []Tag{{Key: "key", Value: "value"},{Key: "longerkey", Value: "value2"}})
	return fwfTable
}

func TestPrint(t *testing.T) {
	var data = []struct {
		withHeader bool
		withTags   bool
		expectedOutputFile string
	} {
		{false, false, "noHeadNoTags"},
		{true, false, "headNoTags"},
		{true, true, "headAndTags"},
		{false, true, "noHeadWithTags"},
	}
	for _, d := range data {
		t.Run(d.expectedOutputFile, func(t *testing.T) {
			var output bytes.Buffer
			createTestTable().Print(&output, d.withHeader, d.withTags)
			bytes, err := ioutil.ReadFile(filepath.Join("testdata", d.expectedOutputFile))
			if err != nil {
				t.Fatalf("error reading test resource %s: %v", d.expectedOutputFile, err)
			}
			if output.String() != string(bytes) {
				t.Errorf("output got:\n%s, want:\n%s", output.String(), string(bytes))
			}
		})
	}
}