package main

import (
	"github.com/deepalert/deepalert"
	"github.com/google/go-github/v27/github"
	"github.com/m-mizutani/golambda"
)

type GithubSettings githubSettings
type Arguments arguments

func Publish(report deepalert.Report, settings GithubSettings) (*github.Issue, error) {
	return publishToGithub(report, githubSettings(settings))
}

var (
	ReportToBody = reportToBody
)

func Handler(args Arguments, event golambda.Event) error {
	return handler(arguments(args), event)
}
