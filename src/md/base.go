package md

import (
	"fmt"
	"io"
)

type Node interface {
	Render(w io.Writer) error
}

type Container struct {
	Children []Node
}

func (x *Container) Append(node Node) {
	x.Children = append(x.Children, node)
}

func (x *Container) Extend(nodes []Node) {
	x.Children = append(x.Children, nodes...)
}

func (x *Container) Render(w io.Writer) error {
	for _, node := range x.Children {
		if err := node.Render(w); err != nil {
			return err
		}
	}

	return nil
}

type Literal string

func ToLiteral(s string) *Literal {
	p := Literal(s)
	return &p
}

func ToLiteralf(f string, v ...interface{}) *Literal {
	p := Literal(fmt.Sprintf(f, v...))
	return &p
}

func (x *Literal) Render(w io.Writer) error {
	if _, err := w.Write([]byte(*x)); err != nil {
		return err
	}

	return nil
}

type Contents []Node

func (x Contents) Render(w io.Writer) error {
	for _, node := range x {
		if err := node.Render(w); err != nil {
			return err
		}
	}

	return nil
}
