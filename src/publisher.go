package main

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/deepalert/deepalert"
	"github.com/google/go-github/v27/github"
	"github.com/m-mizutani/golambda"
	"github.com/pkg/errors"
)

type githubSettings struct {
	GithubEndpoint   string `json:"github_endpoint"`
	GithubRepo       string `json:"github_repo"`
	GithubAppID      string `json:"github_app_id"`
	GithubInstallID  string `json:"github_install_id"`
	GithubPrivateKey string `json:"github_private_key"`
}

func (x githubSettings) hasAppSettings() bool {
	return (x.GithubAppID != "" && x.GithubInstallID != "" && x.GithubPrivateKey != "")
}

func (x githubSettings) newClient() (*github.Client, error) {
	appID, err := strconv.ParseInt(x.GithubAppID, 10, 64)
	if err != nil {
		return nil, golambda.WrapError(err, "Fail to parse appID: %s", x.GithubAppID)
	}

	installID, err := strconv.ParseInt(x.GithubInstallID, 10, 64)
	if err != nil {
		return nil, golambda.WrapError(err, "Fail to parse InstallID: %s", x.GithubInstallID)
	}

	privateKey, err := base64.StdEncoding.DecodeString(x.GithubPrivateKey)
	if err != nil {
		return nil, golambda.WrapError(err, "Fail to decode privateKey as base64, len:%d", len(x.GithubPrivateKey))
	}

	return newGithubAppClient(x.GithubEndpoint, appID, installID, privateKey)
}

func newGithubAppClient(endpoint string, appID int64, installID int64, privateKey []byte) (*github.Client, error) {
	tr := http.DefaultTransport

	itr, err := ghinstallation.New(tr, appID, installID, privateKey)
	if !strings.HasSuffix(endpoint, "/") {
		endpoint = endpoint + "/"
	}

	if err != nil {
		return nil, errors.Wrap(err, "Fail to create GH client")
	}
	itr.BaseURL = endpoint + "app"

	client := github.NewClient(&http.Client{Transport: itr})

	if endpoint != "" {
		url, err := url.Parse(endpoint)
		if err != nil {
			return nil, golambda.WrapError(err, "Invalid URL format: %s", endpoint)
		}

		client.BaseURL = url
	}

	logger.With("client", client).Trace("Github Client is created")

	return client, nil
}

func reportToTitle(report deepalert.Report) string {
	return fmt.Sprintf("[%s] %s: %s", report.Alerts[0].Detector, report.Alerts[0].RuleName, report.Alerts[0].Description)
}

func publishToGithub(report deepalert.Report, settings githubSettings) (*github.Issue, error) {
	logger.With("report", report).Info("Publishing report")
	var issue *github.Issue

	client, err := settings.newClient()
	if err != nil {
		return nil, errors.Wrap(err, "Fail to create ")
	}

	switch report.Status {
	case deepalert.StatusNew:
		fallthrough
	case deepalert.StatusMore:
		path, err := publishAlert(client, report, settings)
		if err != nil {
			return nil, errors.Wrap(err, "Fail to publish alert")
		}
		logger.With("path", path).Info("published alert")

	case deepalert.StatusPublished:
		if report.Result.Severity != deepalert.SevSafe {
			issue, err = publishReport(client, report, settings)
			if err != nil {
				return nil, err
			}
			logger.With("issue", issue).Info("publish only a 'published' report")
		} else {
			logger.Info("Report is not published because the severity is safe")
		}
	}

	return issue, nil
}

func reportToPath(report deepalert.Report) string {
	return fmt.Sprintf("%s/%s/", report.CreatedAt.Format("2006/01/02"), report.ID)
}

func publishAlert(client *github.Client, report deepalert.Report, settings githubSettings) (string, error) {
	ctx := context.Background()
	arr := strings.Split(settings.GithubRepo, "/")
	owner := arr[0]
	repo := arr[1]

	for _, alert := range report.Alerts {
		nodes := buildAlert(alert)

		buf := new(bytes.Buffer)
		for _, node := range nodes {
			if err := node.Render(buf); err != nil {
				return "", err
			}
		}

		data := buf.Bytes()
		sha := sha1.Sum(data)
		hv := fmt.Sprintf("%040x", sha)
		opt := github.RepositoryContentFileOptions{
			Message: github.String(fmt.Sprintf("[Alert] %s: %s", alert.RuleName, alert.Description)),
			Content: data,
			SHA:     github.String(hv),
			Branch:  github.String("master"),
		}
		dpath := reportToPath(report)
		fpath := fmt.Sprintf("%s%s_%s.md", dpath,
			alert.Timestamp.Format("20060102_150405"), hv)
		content, resp, err := client.Repositories.CreateFile(ctx, owner, repo, fpath, &opt)
		logger.With("content", content).With("error", err).Info("Create file")
		if resp != nil {
			logger.With("code", resp.StatusCode).Info("Status code")
		}

		if err != nil || resp.StatusCode != 409 {
			return "", errors.Wrap(err, "Fail to create a file")
		}
	}
	return "", nil
}

func publishReport(client *github.Client, report deepalert.Report, settings githubSettings) (*github.Issue, error) {
	title := reportToTitle(report)
	buf, err := reportToBody(report)
	if err != nil {
		return nil, errors.Wrap(err, "Fail to build body")
	}
	body := buf.String()

	ctx := context.Background()
	issueReq := github.IssueRequest{
		Title: github.String(title),
		Body:  github.String(body),
	}
	arr := strings.Split(settings.GithubRepo, "/")
	if len(arr) != 2 {
		return nil, fmt.Errorf("%s is not repository format, must be {owner}/{repo_name}", settings.GithubRepo)
	}

	issue, resp, err := client.Issues.Create(ctx, arr[0], arr[1], &issueReq)
	if err != nil {
		return nil, errors.Wrap(err, "Fail to create issue")
	}
	if resp.StatusCode != 201 {
		return nil, fmt.Errorf("Fail to create issue because response code is not 201: %d", resp.StatusCode)
	}

	return issue, nil
}
