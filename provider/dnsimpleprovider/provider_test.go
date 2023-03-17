package dnsimpleprovider_test

import (
	"context"
	"os"
	"time"

	"github.com/dnsimple/dnsimple-go/dnsimple"
	. "github.com/dogmatiq/proclaim/provider/dnsimpleprovider"
	"github.com/dogmatiq/proclaim/provider/dnsimpleprovider/internal/dnsimplex"
	"github.com/dogmatiq/proclaim/provider/internal/providertest"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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

			deleteAllRecords(client)

			return providertest.TestContext{
				Provider: &Provider{
					API: client,
				},
				Domain:     domain,
				NameServer: nameServer,
			}
		},
	)
})

func deleteAllRecords(client *dnsimple.Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := dnsimplex.Each(
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
	Expect(err).ShouldNot(HaveOccurred())
}
