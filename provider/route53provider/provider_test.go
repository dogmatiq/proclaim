package route53provider_test

import (
	"context"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/dogmatiq/proclaim/provider/internal/providertest"
	. "github.com/dogmatiq/proclaim/provider/route53provider"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	hostedZoneID = "Z06715307D3NWJ3JMTGQ"
	domain       = "proclaim-test.dogmatiq.io"
	nameServer   = "ns-1143.awsdns-14.org"
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

			deleteAllRecords(client)

			return providertest.TestContext{
				Provider: &Provider{
					PartitionID: "aws", // TODO: obtain this dynamically
					Client:      client,
				},
				Domain:     domain,
				NameServer: nameServer,
			}
		},
	)
})

func deleteAllRecords(client *route53.Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cs := &types.ChangeBatch{}

	in := &route53.ListResourceRecordSetsInput{
		HostedZoneId: aws.String(hostedZoneID),
	}

	for {
		out, err := client.ListResourceRecordSets(ctx, in)
		Expect(err).ShouldNot(HaveOccurred())

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
		return
	}

	_, err := client.ChangeResourceRecordSets(
		ctx,
		&route53.ChangeResourceRecordSetsInput{
			ChangeBatch:  cs,
			HostedZoneId: aws.String(hostedZoneID),
		},
	)
	Expect(err).ShouldNot(HaveOccurred())
}
