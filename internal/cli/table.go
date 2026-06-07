package cli

import (
	"fmt"
	"io"
	"strings"
)

type Table struct {
	headers []string
	rows    [][]string
	widths  []int
}

func NewTable(headers ...string) *Table {
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	return &Table{
		headers: headers,
		widths:  widths,
	}
}

func (t *Table) AddRow(cols ...string) {
	for len(cols) < len(t.headers) {
		cols = append(cols, "")
	}
	for i, c := range cols {
		if i < len(t.widths) && len(c) > t.widths[i] {
			t.widths[i] = len(c)
		}
	}
	t.rows = append(t.rows, cols)
}

func (t *Table) Render(w io.Writer) {
	// Header
	for i, h := range t.headers {
		if i > 0 {
			fmt.Fprint(w, "  ")
		}
		fmt.Fprintf(w, "%-*s", t.widths[i], strings.ToUpper(h))
	}
	fmt.Fprintln(w)

	// Rows
	for _, row := range t.rows {
		for i, c := range row {
			if i >= len(t.widths) {
				break
			}
			if i > 0 {
				fmt.Fprint(w, "  ")
			}
			fmt.Fprintf(w, "%-*s", t.widths[i], c)
		}
		fmt.Fprintln(w)
	}
}

func FormatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func Truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
