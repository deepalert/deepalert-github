package main

import (
	"sort"

	"github.com/deepalert/deepalert"
	"github.com/deepalert/deepalert-github/src/md"
)

func buildActivitiesSection(activities []deepalert.EntityActivity) (nodes []md.Node) {
	if len(activities) == 0 {
		return
	}

	table := md.Table{
		Haed: md.TableHead{
			Cols: []md.TableCol{
				{Content: md.ToLiteral("LastSeen")},
				{Content: md.ToLiteral("ServiceName")},
				{Content: md.ToLiteral("RemoteAddr")},
				{Content: md.ToLiteral("Principal")},
				{Content: md.ToLiteral("Action")},
				{Content: md.ToLiteral("Target")},
			},
		},
	}

	sort.Slice(activities, func(i, j int) bool {
		return activities[i].LastSeen.After(activities[j].LastSeen)
	})

	for _, act := range activities {
		table.Rows = append(table.Rows, md.TableRow{
			Cols: []md.TableCol{
				{Content: md.ToLiteral(act.LastSeen.Format(timeFormat))},
				{Content: md.ToLiteral(act.ServiceName)},
				{Content: md.ToLiteral(act.RemoteAddr)},
				{Content: md.ToLiteral(act.Principal)},
				{Content: md.ToLiteral(act.Action)},
				{Content: md.ToLiteral(act.Target)},
			},
		})
	}

	nodes = append(nodes, []md.Node{
		&md.Heading{Level: 3, Content: md.ToLiteral("Activities")},
		&table,
	}...)

	return
}
