package route53provider_test

import (
	"context"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/dogmatiq/proclaim/provider/internal/providertest"
	. "github.com/dogmatiq/proclaim/provider/route53provider"
	"github.com/go-logr/logr"
	"github.com/miekg/dns"
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
				GetRecords: func(ctx context.Context) ([]dns.RR, error) {
					var records []dns.RR

					in := &route53.ListResourceRecordSetsInput{
						HostedZoneId: aws.String(hostedZoneID),
					}

					for {
						out, err := client.ListResourceRecordSets(ctx, in)
						if err != nil {
							return nil, err
						}

						for _, set := range out.ResourceRecordSets {
							records = append(records, convertRecords(set)...)
						}

						if !out.IsTruncated {
							break
						}

						in.StartRecordIdentifier = out.NextRecordIdentifier
						in.StartRecordName = out.NextRecordName
						in.StartRecordType = out.NextRecordType
					}

					return records, nil
				},
				DeleteRecords: func(ctx context.Context, service string) error {
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

							if strings.Contains(*rr.Name, service) {
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

func convertRecords(set types.ResourceRecordSet) []dns.RR {
	var records []dns.RR

	for _, rec := range set.ResourceRecords {
		switch set.Type {
		case types.RRTypePtr:
			records = append(records, &dns.PTR{
				Hdr: dns.RR_Header{
					Name:   *set.Name,
					Rrtype: dns.TypePTR,
					Class:  dns.ClassINET,
					Ttl:    uint32(*set.TTL),
				},
				Ptr: *rec.Value,
			})
		case types.RRTypeSrv:
			parts := strings.Split(*rec.Value, " ")

			priority, err := strconv.Atoi(parts[0])
			if err != nil {
				panic(err)
			}

			weight, err := strconv.Atoi(parts[1])
			if err != nil {
				panic(err)
			}

			port, err := strconv.Atoi(parts[2])
			if err != nil {
				panic(err)
			}

			records = append(records, &dns.SRV{
				Hdr: dns.RR_Header{
					Name:   *set.Name,
					Rrtype: dns.TypeSRV,
					Class:  dns.ClassINET,
					Ttl:    uint32(*set.TTL),
				},
				Target:   parts[3],
				Port:     uint16(port),
				Priority: uint16(priority),
				Weight:   uint16(weight),
			})
		case types.RRTypeTxt:
			records = append(records, &dns.TXT{
				Hdr: dns.RR_Header{
					Name:   *set.Name,
					Rrtype: dns.TypeTXT,
					Class:  dns.ClassINET,
					Ttl:    uint32(*set.TTL),
				},
				Txt: strings.Split(
					strings.Trim(*rec.Value, `"`),
					`" "`,
				),
			})
		}
	}

	return records
}
