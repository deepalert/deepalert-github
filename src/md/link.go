package md

import (
	"fmt"
	"io"
)

type Link struct {
	Content Node
	URL     string
}

func (x *Link) Render(w io.Writer) error {
	if _, err := w.Write([]byte("[")); err != nil {
		return err
	}

	if x.Content != nil {
		if err := x.Content.Render(w); err != nil {
			return err
		}
	}

	if _, err := w.Write([]byte(fmt.Sprintf("](%s)", x.URL))); err != nil {
		return err
	}

	return nil
}
