package main

import (
	"fmt"

	"github.com/deepalert/deepalert"
	"github.com/deepalert/deepalert-github/src/md"
)

func buildHostInspections(hosts []*deepalert.ContentHost,
	attr deepalert.Attribute) (nodes []md.Node) {

	if len(hosts) == 0 {
		return
	}

	for _, host := range hosts {
		nodes = append(nodes, &md.Heading{
			Level:   2,
			Content: md.ToLiteral(fmt.Sprintf("Host: %s", attr.Value)),
		})

		nodes = append(nodes, buildReportHostBaseSection(host)...)
		nodes = append(nodes, buildActivitiesSection(host.Activities)...)
		nodes = append(nodes, buildReportHostDomainSection(host.RelatedDomains)...)
		nodes = append(nodes, buildReportHostURLSection(host.RelatedURLs)...)
		nodes = append(nodes, buildReportHostMalwareSection(host.RelatedMalware)...)
		nodes = append(nodes, buildReportHostSoftwareSection(host.Software)...)

		if len(nodes) == 1 {
			nodes = append(nodes, md.ToLiteral("N/A"))
		}
	}

	return
}

func mergeReportHost(contents []deepalert.ContentHost) (merged deepalert.ContentHost) {
	for _, c := range contents {
		merged.IPAddr = append(merged.IPAddr, c.IPAddr...)
		merged.Country = append(merged.Country, c.Country...)
		merged.ASOwner = append(merged.ASOwner, c.ASOwner...)
		merged.UserName = append(merged.UserName, c.UserName...)
		merged.Owner = append(merged.Owner, c.Owner...)
		merged.OS = append(merged.OS, c.OS...)
		merged.MACAddr = append(merged.MACAddr, c.MACAddr...)
		merged.HostName = append(merged.HostName, c.HostName...)

		merged.Activities = append(merged.Activities, c.Activities...)
		merged.RelatedDomains = append(merged.RelatedDomains, c.RelatedDomains...)
		merged.RelatedURLs = append(merged.RelatedURLs, c.RelatedURLs...)
		merged.RelatedMalware = append(merged.RelatedMalware, c.RelatedMalware...)
	}

	return
}

func buildReportHostBaseSection(merged *deepalert.ContentHost) []md.Node {
	type itemSet struct {
		title string
		items []string
	}
	targets := []itemSet{
		{title: "IPAddr: ", items: merged.IPAddr},
		{title: "Country: ", items: merged.Country},
		{title: "ASOwner: ", items: merged.ASOwner},
		{title: "UserName: ", items: merged.UserName},
		{title: "Owner: ", items: merged.Owner},
		{title: "OS: ", items: merged.OS},
		{title: "MACAddr: ", items: merged.MACAddr},
		{title: "HostName: ", items: merged.HostName},
	}

	list := md.List{}
	for _, target := range targets {
		if len(target.items) > 0 {
			listContents := md.Contents{md.ToLiteral(target.title)}
			listContents = append(listContents, joinAsCode(target.items)...)

			list.Items = append(list.Items, md.ListItem{Content: listContents})
		}
	}

	return []md.Node{&list}
}

func buildReportHostActivitiesSection(activities []deepalert.EntityActivity) (nodes []md.Node) {
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

func buildReportHostDomainSection(activities []deepalert.EntityDomain) (nodes []md.Node) {
	if len(activities) == 0 {
		return
	}

	table := md.Table{
		Haed: md.TableHead{
			Cols: []md.TableCol{
				{Content: md.ToLiteral("Timestamp")},
				{Content: md.ToLiteral("Name")},
				{Content: md.ToLiteral("Source")},
			},
		},
	}

	for _, act := range activities {
		table.Rows = append(table.Rows, md.TableRow{
			Cols: []md.TableCol{
				{Content: md.ToLiteral(act.Timestamp.Format(timeFormat))},
				{Content: md.ToLiteral(act.Name)},
				{Content: md.ToLiteral(act.Source)},
			},
		})
	}

	nodes = append(nodes, []md.Node{
		&md.Heading{Level: 3, Content: md.ToLiteral("Related Domains")},
		&table,
	}...)

	return
}

func buildReportHostURLSection(activities []deepalert.EntityURL) (nodes []md.Node) {
	if len(activities) == 0 {
		return
	}

	table := md.Table{
		Haed: md.TableHead{
			Cols: []md.TableCol{
				{Content: md.ToLiteral("Timestamp")},
				{Content: md.ToLiteral("URL")},
				{Content: md.ToLiteral("Reference")},
				{Content: md.ToLiteral("Source")},
			},
		},
	}

	for _, act := range activities {
		table.Rows = append(table.Rows, md.TableRow{
			Cols: []md.TableCol{
				{Content: md.ToLiteral(act.Timestamp.Format(timeFormat))},
				{Content: md.ToLiteral(act.URL)},
				{Content: md.ToLiteral(act.Reference)},
				{Content: md.ToLiteral(act.Source)},
			},
		})
	}

	nodes = append(nodes, []md.Node{
		&md.Heading{Level: 3, Content: md.ToLiteral("Related URLs")},
		&table,
	}...)

	return
}

func buildReportHostMalwareSection(malware []deepalert.EntityMalware) (nodes []md.Node) {
	if len(malware) == 0 {
		return
	}

	var venders []string
	venderMap := map[string]struct{}{}
	for _, entity := range malware {
		for _, scan := range entity.Scans {
			venderMap[scan.Vendor] = struct{}{}
		}
	}
	for vender := range venderMap {
		venders = append(venders, vender)
	}

	// Build table head entities
	table := md.Table{
		Haed: md.TableHead{
			Cols: []md.TableCol{
				{Content: md.ToLiteral("Timestamp")},
				{Content: md.ToLiteral("Relation")},
			},
		},
	}
	for _, vender := range venders {
		table.Haed.Cols = append(table.Haed.Cols, md.TableCol{
			Content: md.ToLiteral(vender),
			Align:   md.AlignCenter,
		})
	}

	// Build table body entities
	empty := md.TableCol{}
	for _, act := range malware {
		row := md.TableRow{
			Cols: []md.TableCol{
				{Content: md.ToLiteral(act.Timestamp.Format(timeFormat))},
				{Content: md.ToLiteral(act.Relation)},
			},
		}
		for _, vendor := range venders {
			var col *md.TableCol
			for _, scan := range act.Scans {
				if scan.Vendor == vendor {
					col = &md.TableCol{Content: md.ToLiteral(scan.Name)}
					break
				}
			}

			if col != nil {
				row.Cols = append(row.Cols, *col)
			} else {
				row.Cols = append(row.Cols, empty)
			}
		}

		table.Rows = append(table.Rows, row)
	}

	nodes = append(nodes, []md.Node{
		&md.Heading{Level: 3, Content: md.ToLiteral("Related Malware")},
		&table,
	}...)

	return
}

func buildReportHostSoftwareSection(activities []deepalert.EntitySoftware) (nodes []md.Node) {
	if len(activities) == 0 {
		return
	}

	table := md.Table{
		Haed: md.TableHead{
			Cols: []md.TableCol{
				{Content: md.ToLiteral("LastSeen")},
				{Content: md.ToLiteral("Name")},
				{Content: md.ToLiteral("Location")},
			},
		},
	}

	for _, act := range activities {
		table.Rows = append(table.Rows, md.TableRow{
			Cols: []md.TableCol{
				{Content: md.ToLiteral(act.LastSeen.Format(timeFormat))},
				{Content: md.ToLiteral(act.Name)},
				{Content: md.ToLiteral(act.Location)},
			},
		})
	}

	nodes = append(nodes, []md.Node{
		&md.Heading{Level: 3, Content: md.ToLiteral("Installed Software")},
		&table,
	}...)

	return
}
