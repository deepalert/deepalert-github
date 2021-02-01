package main

import (
	"fmt"

	"github.com/deepalert/deepalert"
	"github.com/deepalert/deepalert-github/src/md"
)

func buildUserInspections(users []*deepalert.ContentUser,
	attr deepalert.Attribute) (nodes []md.Node) {

	for _, user := range users {
		nodes = append(nodes, &md.Heading{
			Level:   2,
			Content: md.ToLiteral(fmt.Sprintf("User: `%s`", attr.Value)),
		})

		nodes = append(nodes, buildActivitiesSection(user.Activities)...)

		if len(nodes) == 1 {
			nodes = append(nodes, md.ToLiteral("N/A"))
		}
	}

	return
}
