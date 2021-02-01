package md

import "io"

type HorizontalRules struct{}

func (x *HorizontalRules) Render(w io.Writer) error {
	if _, err := w.Write([]byte("------\n\n")); err != nil {
		return err
	}

	return nil
}
