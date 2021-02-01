package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/deepalert/deepalert"
	"github.com/deepalert/deepalert-github/src/md"
)

const timeFormat = "2006-01-02 15:04"

func attrToContents(attr *deepalert.Attribute) md.Contents {
	nodes := []md.Node{
		md.ToLiteral(fmt.Sprintf("%s", attr.Key)),
	}

	switch attr.Type {
	case "":
		nodes = append(nodes, []md.Node{
			md.ToLiteral(": "),
			md.ToCode(attr.Value),
		}...)

	case "json":
		var ppJSON bytes.Buffer
		var jdata string
		err := json.Indent(&ppJSON, []byte(attr.Value), "", "  ")
		if err != nil {
			jdata = attr.Value
		} else {
			jdata = ppJSON.String()
		}

		nodes = append(nodes, []md.Node{
			md.ToLiteral(" ("),
			md.ToCode(string(attr.Type)),
			md.ToLiteral("): \n"),
			md.ToCodeBlock(jdata),
			md.ToLiteral("\n"),
		}...)

	case deepalert.TypeURL:
		if attr.Context.Have(deepalert.CtxAdditionalInfo) {
			nodes = append(nodes, []md.Node{
				md.ToLiteral(": "),
				&md.Link{
					Content: md.ToLiteral(attr.Value),
					URL:     attr.Value,
				},
			}...)
		} else {
			nodes = append(nodes, md.ToCode(attr.Value))
		}

	default:
		nodes = append(nodes, []md.Node{
			md.ToLiteral(" ("),
			md.ToCode(string(attr.Type)),
			md.ToLiteral("): "),
			md.ToCode(attr.Value),
		}...)
	}

	return md.Contents(nodes)
}

func buildSummary(report deepalert.Report) []md.Node {
	attrList := &md.List{}

	for _, attr := range report.Attributes {
		attrList.Items = append(attrList.Items, md.ListItem{
			Content: attrToContents(attr),
		})
	}

	nodes := []md.Node{
		&md.Heading{
			Level:   1,
			Content: md.ToLiteral("Summary"),
		},
		&md.List{
			Items: []md.ListItem{
				{Content: md.Contents{
					md.ToLiteral("Severity: "),
					md.ToBold(string(report.Result.Severity)),
				}},
				{Content: md.Contents{
					md.ToLiteral("Reason: "),
					md.ToLiteral(report.Result.Reason),
				}},
				{Content: md.Contents{
					md.ToLiteral("Detected by "),
					md.ToCode(report.Alerts[0].Detector),
				}},
				{Content: md.Contents{
					md.ToLiteral("Rule: "),
					md.ToCode(report.Alerts[0].RuleName),
				}},
				{Content: md.Contents{
					md.ToLiteral("Created at: " + report.CreatedAt.String()),
				}},
				{Content: md.Contents{
					md.ToLiteralf("Alert reports: [link](../tree/master/%s)", reportToPath(report)),
				}},
			},
		},
		&md.Heading{
			Level:   2,
			Content: md.ToLiteral("Attributes"),
		},
		attrList,
	}

	return nodes
}

func joinAsCode(ss []string) []md.Node {
	var nodes []md.Node
	for i, s := range ss {
		nodes = append(nodes, md.ToCode(s))
		if i+1 < len(ss) {
			nodes = append(nodes, md.ToLiteral(", "))
		}
	}
	return nodes
}

func buildInspections(report deepalert.Report) []md.Node {
	nodes := []md.Node{
		&md.Heading{
			Level:   1,
			Content: md.ToLiteral("Inspection Reports"),
		},
	}

	for _, section := range report.Sections {
		nodes = append(nodes, buildHostInspections(section.Hosts, section.Attr)...)
		nodes = append(nodes, buildUserInspections(section.Users, section.Attr)...)
		nodes = append(nodes, buildBinaryInspections(section.Binaries, section.Attr)...)
	}

	return nodes
}

func buildAlert(alert *deepalert.Alert) []md.Node {
	title := fmt.Sprintf("[%s] %s: %s", alert.Detector, alert.RuleName, alert.Description)

	nodes := []md.Node{
		&md.Heading{
			Level:   1,
			Content: md.ToLiteral(title),
		},
	}

	nodes = append(nodes, []md.Node{
		&md.List{
			Items: []md.ListItem{
				{Content: md.Contents{
					md.ToLiteral("Detector: "),
					md.ToCode(alert.Detector),
				}},
				{Content: md.Contents{
					md.ToLiteral("RuleName: "),
					md.ToCode(alert.RuleName),
				}},
				{Content: md.Contents{
					md.ToLiteral("RuleID: "),
					md.ToCode(alert.RuleID),
				}},
				{Content: md.Contents{
					md.ToLiteral("AlertKey: "),
					md.ToCode(alert.AlertKey),
				}},
				{Content: md.Contents{
					md.ToLiteral("Description: "),
					md.ToLiteral(alert.Description),
				}},
				{Content: md.Contents{
					md.ToLiteral("Detected at: "),
					md.ToCode(alert.Timestamp.Format(timeFormat)),
				}},
			},
		},
		&md.Heading{Content: md.ToLiteral("Attributes"), Level: 2},
	}...)

	attrList := &md.List{}
	for _, attr := range alert.Attributes {
		attrList.Items = append(attrList.Items, md.ListItem{
			Content: attrToContents(&attr),
		})
	}
	nodes = append(nodes, attrList)

	return nodes
}

func buildSystemReport(report deepalert.Report) (nodes []md.Node) {
	nodes = append(nodes, []md.Node{
		&md.Heading{Content: md.ToLiteral("System Info")},
		&md.List{
			Items: []md.ListItem{
				{
					Content: md.Contents{
						md.ToLiteral("ReportID: "),
						md.ToCode(string(report.ID)),
					},
				},
				{
					Content: md.Contents{
						md.ToLiteral("Status: "),
						md.ToCode(string(report.Status)),
					},
				},
			},
		},
	}...)

	logger.With("nodes", nodes).Info("Built system report")

	return
}

func reportToBody(report deepalert.Report) (*bytes.Buffer, error) {
	doc := &md.Document{}
	doc.Extend(buildSummary(report))
	doc.Extend(buildInspections(report))
	doc.Extend(buildSystemReport(report))

	buf := new(bytes.Buffer)
	if err := doc.Render(buf); err != nil {
		return nil, nil
	}

	return buf, nil
}
