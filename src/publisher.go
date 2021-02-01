package main

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/deepalert/deepalert"
	"github.com/google/go-github/v27/github"
	"github.com/m-mizutani/golambda"
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
		return nil, golambda.WrapError(err, "Fail to parse appID").With("apID", x.GithubAppID)
	}

	installID, err := strconv.ParseInt(x.GithubInstallID, 10, 64)
	if err != nil {
		return nil, golambda.WrapError(err, "Fail to parse InstallID").With("installID", x.GithubInstallID)
	}

	privateKey, err := base64.StdEncoding.DecodeString(x.GithubPrivateKey)
	if err != nil {
		return nil, golambda.WrapError(err, "Fail to decode privateKey as base64").With("len", len(x.GithubPrivateKey))
	}

	return newGithubAppClient(x.GithubEndpoint, appID, installID, privateKey)
}

func newGithubAppClient(endpoint string, appID int64, installID int64, privateKey []byte) (*github.Client, error) {
	tr := http.DefaultTransport

	logger.With("appID", appID).
		With("endpoint", endpoint).
		With("installID", installID).
		With("privateKey.length", len(privateKey)).
		Debug("Creating github app client")

	itr, err := ghinstallation.New(tr, appID, installID+1000, privateKey)
	if err != nil {
		return nil, golambda.WrapError(err, "Fail to create GH client").With("appID", appID).With("installID", installID)
	}

	var client *github.Client
	if endpoint == "" {
		client = github.NewClient(&http.Client{Transport: itr})
	} else {
		itr.BaseURL = strings.TrimLeft(endpoint, "/")
		client, err = github.NewEnterpriseClient(endpoint, endpoint, &http.Client{Transport: itr})
		if err != nil {
			return nil, golambda.WrapError(err).With("endpoint", endpoint)
		}
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
		return nil, err
	}

	switch report.Status {
	case deepalert.StatusNew:
		fallthrough
	case deepalert.StatusMore:
		path, err := publishAlert(client, report, settings)
		if err != nil {
			return nil, err
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
		if err != nil {
			if strings.Contains(err.Error(), ": 409 ") {
				logger.With("owner", arr[0]).
					With("repo", arr[1]).
					With("content", content).
					With("fpath", fpath).Info("409 error (conflicted) is returned, but ignore")
				return "", nil
			}

			e := golambda.NewError("Failed to create a file").
				With("owner", arr[0]).
				With("repo", arr[1]).
				With("content", content).
				With("fpath", fpath)
			if resp != nil {
				e = e.With("code", resp.StatusCode)
				if body, err := ioutil.ReadAll(resp.Body); err != nil {
					e = e.With("read error", err)
				} else {
					e = e.With("body", body)
				}
			}
			return "", e
		}
	}
	return "", nil
}

func publishReport(client *github.Client, report deepalert.Report, settings githubSettings) (*github.Issue, error) {
	title := reportToTitle(report)
	buf, err := reportToBody(report)
	if err != nil {
		return nil, err
	}
	body := buf.String()

	ctx := context.Background()
	issueReq := github.IssueRequest{
		Title: github.String(title),
		Body:  github.String(body),
	}
	arr := strings.Split(settings.GithubRepo, "/")
	if len(arr) != 2 {
		return nil, golambda.NewError("invalid repository format, must be {owner}/{repo_name}").With("repo", settings.GithubRepo)
	}

	issue, resp, err := client.Issues.Create(ctx, arr[0], arr[1], &issueReq)
	if err != nil {
		e := golambda.NewError("Failed to create an issue").
			With("owner", arr[0]).
			With("repo", arr[1])
		if resp != nil {
			e = e.With("code", resp.StatusCode)
			if body, err := ioutil.ReadAll(resp.Body); err != nil {
				e = e.With("read error", err)
			} else {
				e = e.With("body", body)
			}
		}
		return nil, e
	}

	if resp.StatusCode != 201 {
		return nil, golambda.NewError("Fail to create issue because response code is not 201").With("code", resp.StatusCode)
	}

	return issue, nil
}
