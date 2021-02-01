package main

import (
	"github.com/Netflix/go-env"
	"github.com/deepalert/deepalert"
	"github.com/m-mizutani/golambda"
)

var logger = golambda.Logger

type arguments struct {
	SecretARN      string `env:"SECRET_ARN"`
	GitHubEndpoint string `env:"GITHUB_ENDPOINT"`
	GitHubRepo     string `env:"GITHUB_REPO"`

	NewSM golambda.SecretsManagerFactory
}

func handler(args arguments, event golambda.Event) error {
	records, err := event.DecapSNSonSQSMessage()
	if err != nil {
		return err
	}

	for _, record := range records {
		var report deepalert.Report
		if err := record.Bind(&report); err != nil {
			return err
		}

		var settings githubSettings
		if err := golambda.GetSecretValuesWithFactory(args.SecretARN, &settings, args.NewSM); err != nil {
			return err
		}

		settings.GithubEndpoint = args.GitHubEndpoint
		settings.GithubRepo = args.GitHubRepo

		if _, err := publishToGithub(report, settings); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	golambda.Start(func(event golambda.Event) (interface{}, error) {
		var args arguments
		if _, err := env.UnmarshalFromEnviron(&args); err != nil {
			return nil, golambda.WrapError(err, "Failed to unmarshal env vars")
		}

		if err := handler(args, event); err != nil {
			return nil, err
		}
		return nil, nil
	})
}
