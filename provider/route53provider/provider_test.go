package route53provider_test

import (
	"context"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/dogmatiq/proclaim/provider/internal/providertest"
	. "github.com/dogmatiq/proclaim/provider/route53provider"
)

func TestProvider(t *testing.T) {
	if os.Getenv("DOGMATIQ_TEST_AWS_ACCESS_KEY_ID") == "" {
		t.Skip("DOGMATIQ_TEST_AWS_ACCESS_KEY_ID is not defined")
	}

	client := route53.NewFromConfig(
		aws.Config{
			Region: "us-east-1", // Irrelevant for Route 53, but required by the SDK.
			Credentials: aws.CredentialsProviderFunc(
				func(ctx context.Context) (aws.Credentials, error) {
					return aws.Credentials{
						AccessKeyID:     os.Getenv("DOGMATIQ_TEST_AWS_ACCESS_KEY_ID"),
						SecretAccessKey: os.Getenv("DOGMATIQ_TEST_AWS_SECRET_ACCESS_KEY"),
					}, nil
				},
			),
			// Allow a large number of retries because we often hit rate
			// limits in CI when a PR is submitted because there are
			// concurrent executions of the tests for both the branch
			// and the PR itself.
			RetryMaxAttempts: 100,
		},
	)

	providertest.Run(
		t,
		providertest.TestContext{
			Provider: &Provider{
				Client: client,
			},
			Domain: "dissolve-test.dogmatiq.io",
		},
	)
}
