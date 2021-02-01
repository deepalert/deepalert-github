package main_test

import (
	"testing"
	"time"

	"github.com/Netflix/go-env"
	"github.com/deepalert/deepalert"
	"github.com/google/uuid"
	"github.com/m-mizutani/golambda"
	"github.com/stretchr/testify/require"

	main "github.com/deepalert/deepalert-github/src"
)

func TestIntegration(t *testing.T) {

	reportID := deepalert.ReportID(uuid.New().String())
	report := deepalert.Report{
		Status: deepalert.StatusPublished,
		Result: deepalert.ReportResult{
			Severity: deepalert.SevUnclassified,
			Reason:   "It's test",
		},
		ID: reportID,
		Alerts: []*deepalert.Alert{
			{
				Detector:    "blue",
				RuleName:    "orange",
				AlertKey:    "five",
				Description: "not sane",
				Timestamp:   time.Now(),
				Attributes: []deepalert.Attribute{
					{
						Type:    deepalert.TypeIPAddr,
						Key:     "source",
						Value:   "192.168.0.1",
						Context: []deepalert.AttrContext{deepalert.CtxRemote},
					},
				},
			},
			{
				Detector:    "blue",
				RuleName:    "orange",
				AlertKey:    "five",
				Description: "timeless",
				Timestamp:   time.Now(),
				Attributes: []deepalert.Attribute{
					{
						Type:    deepalert.TypeIPAddr,
						Key:     "source",
						Value:   "192.168.0.1",
						Context: []deepalert.AttrContext{deepalert.CtxRemote},
					},
				},
			},
		},
		Attributes: []*deepalert.Attribute{
			{
				Type:    deepalert.TypeIPAddr,
				Key:     "source",
				Value:   "192.168.0.1",
				Context: []deepalert.AttrContext{deepalert.CtxRemote},
			},
		},
		Sections: []*deepalert.Section{
			{
				Attr: deepalert.Attribute{
					Type:    deepalert.TypeIPAddr,
					Key:     "source",
					Value:   "192.168.0.1",
					Context: []deepalert.AttrContext{deepalert.CtxRemote},
				},
				Hosts: []*deepalert.ContentHost{
					{
						RelatedDomains: []deepalert.EntityDomain{
							{
								Name:      "example.com",
								Timestamp: time.Now(),
								Source:    "tester",
							},
						},
					},
					{
						IPAddr: []string{"10.0.1.2"},
						RelatedDomains: []deepalert.EntityDomain{
							{
								Name:      "example.net",
								Timestamp: time.Now(),
								Source:    "tester",
							},
						},
					},
				},
			},
			{
				Attr: deepalert.Attribute{
					Type:    deepalert.TypeIPAddr,
					Key:     "source",
					Value:   "192.168.0.2",
					Context: []deepalert.AttrContext{deepalert.CtxRemote},
				},
				Hosts: []*deepalert.ContentHost{
					{
						RelatedMalware: []deepalert.EntityMalware{
							{
								SHA256:    "abcdefg",
								Timestamp: time.Now(),
								Scans: []deepalert.EntityMalwareScan{
									{
										Vendor: "normalVender",
										Name:   "some_malware",
									},
									{
										Vendor: "superVender",
										Name:   "some_malware2",
									},
								},
							},
						},
					},
				},
			},
			{
				Attr: deepalert.Attribute{
					Type:    deepalert.TypeUserName,
					Key:     "name",
					Value:   "blue",
					Context: []deepalert.AttrContext{deepalert.CtxRemote},
				},
				Users: []*deepalert.ContentUser{
					{
						Activities: []deepalert.EntityActivity{
							{
								ServiceName: "magic",
								RemoteAddr:  "10.2.3.4",
							},
						},
					},
				},
			},
		},
	}

	args := main.Arguments{}
	_, err := env.UnmarshalFromEnviron(&args)
	require.NoError(t, err)
	if args.SecretARN == "" || args.GitHubEndpoint == "" || args.GitHubRepo == "" {
		t.Skipf("not enough arguments %v", args)
	}

	var event golambda.Event
	require.NoError(t, event.EncapSNSonSQSMessage(report))
	require.NoError(t, main.Handler(args, event))
}
