package table

import (
	"fmt"
	"io"
)

type Tag struct {
	Key   string
	Value string
}

type FixedWidthFont struct {
	Header []string
	Rows   [][]string
	Tags   [][]Tag
	widths []int
	maxTagKeyLength int
}

func (fwf *FixedWidthFont) updateWidths(row []string) {
	for i, cell := range row {
		if len(cell) > fwf.widths[i] {
			fwf.widths[i] = len(cell)
		}
	}
}

func (fwf *FixedWidthFont) updateMaxTagKeyLength(tags []Tag) {
	for _, tag := range tags {
		if len(tag.Key) > fwf.maxTagKeyLength {
			fwf.maxTagKeyLength = len(tag.Key)
		}
	}
}

func (fwf *FixedWidthFont) AddRow(row []string, tags []Tag) error {
	if len(row) != len(fwf.Header) {
		return Error{Message: fmt.Sprintf("bad row: expected %d, got %d", len(fwf.Header), len(row))}
	}
	fwf.Rows = append(fwf.Rows, row)
	fwf.updateWidths(row)
	fwf.Tags = append(fwf.Tags, tags)
	fwf.updateMaxTagKeyLength(tags)
	return nil
}

func printRow(w io.Writer, formatTokens []string, row []string) {
	for i, cell := range row {
		fmt.Fprintf(w, formatTokens[i], cell)
		if i + 1 < len(row) {
			fmt.Fprint(w, " ")
		}
	}
	fmt.Fprintln(w)
}

func printTags(w io.Writer, format string, tags []Tag, isLastRow bool) {
	for _, tag := range tags {
		fmt.Fprintf(w, format, tag.Key, tag.Value)
	}
	if !isLastRow {
		fmt.Fprintln(w)
	}
}

func (fwf FixedWidthFont) Print(w io.Writer, withHeader bool, withTags bool) {
	var formatTokens = make([]string, 0, len(fwf.widths))
	for _, width := range fwf.widths {
		formatTokens = append(formatTokens, fmt.Sprintf("%%-%ds", width))
	}
	if withHeader {
		printRow(w, formatTokens, fwf.Header)
		fmt.Fprintln(w)
	}
	formatTags := fmt.Sprintf("%%%ds %%s\n", fwf.maxTagKeyLength)
	for i, row := range fwf.Rows {
		printRow(w, formatTokens, row)
		if withTags {
			printTags(w, formatTags, fwf.Tags[i], i + 1 == len(fwf.Rows))
		}
	}
}

func New(headings []string) FixedWidthFont {
	var t = FixedWidthFont{
		Header: headings,
		widths: make([]int, len(headings)),
		Rows:   make([][]string, 0, 10),
		Tags:   make([][]Tag, 0, 10),
	}
	t.updateWidths(headings)
	return t
}

type Error struct {
	Message string
}

func (e Error) Error() string {
	return e.Message
}
