package md

import "io"

type Table struct {
	Haed TableHead
	Rows []TableRow
}

type ColumnAlign int

const (
	AlignLeft ColumnAlign = iota
	AlignCenter
	AlignRight
)

func (x *Table) Render(w io.Writer) error {
	if err := x.Haed.Render(w); err != nil {
		return err
	}

	for _, row := range x.Rows {
		if err := row.Render(w); err != nil {
			return err
		}
	}

	if _, err := w.Write([]byte("\n")); err != nil {
		return err
	}

	return nil
}

type TableHead struct {
	Cols []TableCol
}

func (x *TableHead) Render(w io.Writer) error {
	w.Write([]byte("|"))
	for _, col := range x.Cols {
		w.Write([]byte(" "))
		if err := col.Render(w); err != nil {
			return err
		}
		w.Write([]byte(" |"))
	}
	w.Write([]byte("\n"))

	w.Write([]byte("|"))
	for _, col := range x.Cols {
		var sep string
		switch col.Align {
		case AlignLeft:
			sep = ":------|"
		case AlignCenter:
			sep = ":------:|"
		case AlignRight:
			sep = "-------:|"
		}

		if _, err := w.Write([]byte(sep)); err != nil {
			return err
		}
	}

	if _, err := w.Write([]byte("\n")); err != nil {
		return err
	}

	return nil
}

type TableRow struct {
	Cols []TableCol
}

func (x *TableRow) Render(w io.Writer) error {
	w.Write([]byte("|"))
	for _, col := range x.Cols {
		w.Write([]byte(" "))
		if err := col.Render(w); err != nil {
			return err
		}
		w.Write([]byte(" |"))
	}
	w.Write([]byte("\n"))

	return nil
}

type TableCol struct {
	Align   ColumnAlign
	Content Node
}

func (x *TableCol) Render(w io.Writer) error {
	if x.Content == nil {
		return nil
	}
	return x.Content.Render(w)
}
