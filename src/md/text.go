package md

import (
	"fmt"
	"io"
)

type Code string

func ToCode(s string) *Code {
	code := Code(s)
	return &code
}

func (x *Code) Render(w io.Writer) error {
	c := fmt.Sprintf(" `%s` ", *x)
	if _, err := w.Write([]byte(c)); err != nil {
		return err
	}

	return nil
}

type Bold string

func ToBold(s string) *Bold {
	b := Bold(s)
	return &b
}

func (x *Bold) Render(w io.Writer) error {
	c := fmt.Sprintf(" **%s** ", *x)
	if _, err := w.Write([]byte(c)); err != nil {
		return err
	}

	return nil
}

type Italic string

func ToItalic(s string) *Italic {
	b := Italic(s)
	return &b
}

func (x *Italic) Render(w io.Writer) error {
	c := fmt.Sprintf(" *%s* ", *x)
	if _, err := w.Write([]byte(c)); err != nil {
		return err
	}

	return nil
}

type CodeBlock string

func ToCodeBlock(s string) *CodeBlock {
	b := CodeBlock(s)
	return &b
}

func (x *CodeBlock) Render(w io.Writer) error {
	c := fmt.Sprintf("\n```json\n%s\n```\n", *x)
	if _, err := w.Write([]byte(c)); err != nil {
		return err
	}

	return nil
}
