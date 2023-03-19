package route53provider_test

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/dogmatiq/proclaim/provider/internal/providertest"
	. "github.com/dogmatiq/proclaim/provider/route53provider"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
)

const (
	hostedZoneID = "Z06715307D3NWJ3JMTGQ"
	domain       = "proclaim-test.dogmatiq.io"
	region       = "us-east-1"
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
				},
			)

			return providertest.TestContext{
				Provider: &Provider{
					Client: client,
					Logger: logr.Discard(),
				},
				Domain: domain,
				NameServers: func(ctx context.Context) ([]string, error) {
					var servers []string

					in := &route53.ListResourceRecordSetsInput{
						HostedZoneId: aws.String(hostedZoneID),
					}

					for {
						out, err := client.ListResourceRecordSets(ctx, in)
						if err != nil {
							return nil, err
						}

						for _, rr := range out.ResourceRecordSets {
							if rr.Type == types.RRTypeNs {
								for _, r := range rr.ResourceRecords {
									servers = append(servers, *r.Value)
								}
							}
						}

						if !out.IsTruncated {
							return servers, nil
						}

						in.StartRecordIdentifier = out.NextRecordIdentifier
						in.StartRecordName = out.NextRecordName
						in.StartRecordType = out.NextRecordType
					}
				},
				DeleteRecords: func(ctx context.Context) error {
					cs := &types.ChangeBatch{}

					in := &route53.ListResourceRecordSetsInput{
						HostedZoneId: aws.String(hostedZoneID),
					}

					for {
						out, err := client.ListResourceRecordSets(ctx, in)
						if err != nil {
							return err
						}

						for _, rr := range out.ResourceRecordSets {
							rr := rr // capture loop variable

							switch rr.Type {
							case types.RRTypeNs, types.RRTypeSoa:
								continue
							default:
								cs.Changes = append(
									cs.Changes,
									types.Change{
										Action:            types.ChangeActionDelete,
										ResourceRecordSet: &rr,
									},
								)
							}
						}

						if !out.IsTruncated {
							break
						}

						in.StartRecordIdentifier = out.NextRecordIdentifier
						in.StartRecordName = out.NextRecordName
						in.StartRecordType = out.NextRecordType
					}

					if len(cs.Changes) == 0 {
						return nil
					}

					_, err := client.ChangeResourceRecordSets(
						ctx,
						&route53.ChangeResourceRecordSetsInput{
							ChangeBatch:  cs,
							HostedZoneId: aws.String(hostedZoneID),
						},
					)

					return err
				},
			}
		},
	)
})
