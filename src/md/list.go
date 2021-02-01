package md

import "io"

type List struct {
	Items []ListItem
}

func (x *List) Render(w io.Writer) error {
	for _, item := range x.Items {
		if err := item.Render(w); err != nil {
			return err
		}
	}

	if _, err := w.Write([]byte("\n")); err != nil {
		return err
	}

	return nil
}

type ListItem struct {
	Indent  int
	Content Node
}

func (x *ListItem) Render(w io.Writer) error {
	var prefix string
	for i := 0; i < x.Indent; i++ {
		prefix = prefix + "  "
	}

	if _, err := w.Write([]byte(prefix + "- ")); err != nil {
		return err
	}

	if x.Content != nil {
		if err := x.Content.Render(w); err != nil {
			return err
		}
	}

	if _, err := w.Write([]byte("\n")); err != nil {
		return err
	}

	return nil
}
