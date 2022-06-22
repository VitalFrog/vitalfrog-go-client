package termtable

import (
	"fmt"
	"io"
)

type TermTable struct {
	w             io.Writer
	columns       []int
	columnDivider string
}

func New(w io.Writer, columnDivider string) *TermTable {
	if columnDivider == "" {
		columnDivider = " "
	}

	return &TermTable{
		w:             w,
		columnDivider: columnDivider,
	}

}

func IntPointer(i int) *int {
	return &i
}

func (tt *TermTable) WriteRowDivider(divider rune) error {
	if len(tt.columns) == 0 {
		return fmt.Errorf("need at least a single column")
	}
	dividerLen := len(tt.columns) * len(tt.columnDivider)
	for _, v := range tt.columns {
		dividerLen += v
	}

	for k := 0; k < dividerLen; k++ {
		_, err := fmt.Fprintf(tt.w, string(divider))
		if err != nil {
			return fmt.Errorf("could not Fprint: %w", err)
		}
	}
	fmt.Print("\n")
	return nil
}
func (tt *TermTable) WriteHeader(row []HeaderField) error {
	maxWidth := 0
	for _, v := range row {
		newWidth := v.Field.Len() + 4
		if newWidth > maxWidth {
			maxWidth = newWidth
		}
	}

	cols := make([]Field, 0)
	for _, v := range row {
		colSize := maxWidth
		if v.Width != nil {
			if *v.Width != colSize {
				colSize = *v.Width
			}
		}
		if colSize < v.Field.Len() {
			colSize = v.Field.Len()
		}
		tt.columns = append(tt.columns, colSize)

		cols = append(cols, v.Field)
	}

	err := tt.WriteRow(cols)
	if err != nil {
		return fmt.Errorf("could not WriteRow: %w", err)
	}

	return nil
}

func (tt *TermTable) WriteRow(row []Field) error {
	if len(tt.columns) <= 0 {
		headerCols := make([]HeaderField, 0)
		for _, v := range row {
			headerCols = append(headerCols, HeaderField{
				Field: v,
			})
		}
		return tt.WriteHeader(headerCols)

	}

	if len(row) > len(tt.columns) {
		return fmt.Errorf("row has more columns than what are defined in table. Not allowed")
	}

	for k, v := range row {
		targetLen := tt.columns[k]
		p := ""
		rowLen := v.Len()
		columnDivider := tt.columnDivider
		if k == 0 {
			columnDivider = WhiteSpace(len(tt.columnDivider))
		}
		switch {
		case rowLen > targetLen:
			p = string([]rune(v.String())[:targetLen-4]) + "..."
		case rowLen <= targetLen:
			whiteSpaceCount := targetLen - rowLen
			p = fmt.Sprintf("%s%s%s", columnDivider, v.String(), WhiteSpace(whiteSpaceCount))
		}
		_, err := fmt.Fprint(tt.w, p)
		if err != nil {
			return fmt.Errorf("could not Fprintf: %w", err)
		}
	}
	_, err := fmt.Fprintf(tt.w, "\n")
	if err != nil {
		return fmt.Errorf("could not Fprintf: %w", err)
	}

	return nil
}

func WhiteSpace(x int) string {
	s := ""
	for k := 0; k < x; k++ {
		s += " "
	}
	return s
}
