package md_test

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/deepalert/deepalert-github/src/md"
	"github.com/stretchr/testify/assert"
)

func Test(t *testing.T) {
	buf := new(bytes.Buffer)

	doc := md.Document{}
	doc.Append(&md.Heading{
		Level:   1,
		Content: md.ToLiteral("hoge"),
	})
	doc.Append(&md.List{Items: []md.ListItem{
		md.ListItem{Content: md.ToLiteral("blue")},
		md.ListItem{Content: md.ToLiteral("orange")},
		md.ListItem{Content: md.ToLiteral("magic")},
	}})
	err := doc.Render(buf)

	assert.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "# hoge\n")
	assert.Contains(t, output, "\n- blue\n")
	assert.Contains(t, output, "\n- orange\n")
	assert.Contains(t, output, "\n- magic\n")

	if os.Getenv("VERBOSE") != "" {
		fmt.Println(output)
	}
}
