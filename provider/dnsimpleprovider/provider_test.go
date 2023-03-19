package dnsimpleprovider_test

import (
	"context"
	"os"

	"github.com/dnsimple/dnsimple-go/dnsimple"
	. "github.com/dogmatiq/proclaim/provider/dnsimpleprovider"
	"github.com/dogmatiq/proclaim/provider/dnsimpleprovider/internal/dnsimplex"
	"github.com/dogmatiq/proclaim/provider/internal/providertest"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
)

const (
	accountID  = "12018"
	domain     = "proclaim-test.dogmatiq.io"
	nameServer = "ns3.dnsimple.com"
)

var _ = Describe("type Provider", func() {
	providertest.DeclareTestSuite(
		func(ctx context.Context) providertest.TestContext {
			token := os.Getenv("DOGMATIQ_TEST_DNSIMPLE_TOKEN")
			if token == "" {
				Skip("DOGMATIQ_TEST_DNSIMPLE_TOKEN is not defined")
			}

			client := dnsimple.NewClient(
				dnsimple.StaticTokenHTTPClient(
					ctx,
					token,
				),
			)

			return providertest.TestContext{
				Provider: &Provider{
					Client: client,
					Logger: logr.Discard(),
				},
				Domain: domain,
				NameServers: func(ctx context.Context) ([]string, error) {
					var servers []string
					err := dnsimplex.Each(
						ctx,
						func(opts dnsimple.ListOptions) (*dnsimple.Pagination, []dnsimple.ZoneRecord, error) {
							res, err := client.Zones.ListRecords(
								ctx,
								accountID,
								domain,
								&dnsimple.ZoneRecordListOptions{
									ListOptions: opts,
									Type:        dnsimple.String("NS"),
								},
							)
							if err != nil {
								return nil, nil, err
							}
							return res.Pagination, res.Data, err
						},
						func(rec dnsimple.ZoneRecord) (bool, error) {
							servers = append(servers, rec.Content)
							return true, nil
						},
					)

					return servers, err
				},
				DeleteRecords: func(ctx context.Context) error {
					return dnsimplex.Each(
						ctx,
						func(opts dnsimple.ListOptions) (*dnsimple.Pagination, []dnsimple.ZoneRecord, error) {
							res, err := client.Zones.ListRecords(
								ctx,
								accountID,
								domain,
								&dnsimple.ZoneRecordListOptions{
									ListOptions: opts,
								},
							)
							if err != nil {
								return nil, nil, err
							}
							return res.Pagination, res.Data, err
						},
						func(rec dnsimple.ZoneRecord) (bool, error) {
							switch rec.Type {
							case "NS", "SOA":
								return true, nil
							default:
								_, err := client.Zones.DeleteRecord(
									ctx,
									accountID,
									domain,
									rec.ID,
								)
								return true, err
							}
						},
					)
				},
			}
		},
	)
})
