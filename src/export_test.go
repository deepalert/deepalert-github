package main

import (
	"github.com/deepalert/deepalert"
	"github.com/google/go-github/v27/github"
)

type GithubSettings githubSettings

func Publish(report deepalert.Report, settings GithubSettings) (*github.Issue, error) {
	return publishToGithub(report, githubSettings(settings))
}

var (
	ReportToBody = reportToBody
	Handler      = handler
)
