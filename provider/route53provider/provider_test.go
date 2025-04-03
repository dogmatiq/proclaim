package route53provider_test

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/dogmatiq/proclaim/provider/internal/providertest"
	. "github.com/dogmatiq/proclaim/provider/route53provider"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
)

const (
	domain = "dissolve-test.dogmatiq.io"
	region = "us-east-1"
)

var _ = Describe("type Provider", func() {
	providertest.DeclareTestSuite(
		func(ctx context.Context) providertest.TestContext {
			if os.Getenv("DOGMATIQ_TEST_AWS_ACCESS_KEY_ID") == "" {
				Skip("DOGMATIQ_TEST_AWS_ACCESS_KEY_ID is not defined")
			}

			client := route53.NewFromConfig(
				aws.Config{
					Region: region,
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

			return providertest.TestContext{
				Provider: &Provider{
					Client: client,
					Logger: logr.Discard(),
				},
				Domain: domain,
			}
		},
	)
})
