package dnsimpleprovider_test

import (
	"context"
	"os"
	"strconv"
	"strings"

	"github.com/dnsimple/dnsimple-go/dnsimple"
	. "github.com/dogmatiq/proclaim/provider/dnsimpleprovider"
	"github.com/dogmatiq/proclaim/provider/dnsimpleprovider/internal/dnsimplex"
	"github.com/dogmatiq/proclaim/provider/internal/providertest"
	"github.com/go-logr/logr"
	"github.com/miekg/dns"
	. "github.com/onsi/ginkgo/v2"
)

const (
	accountID  = "12018"
	domain     = "proclaim-test.dogmatiq.io"
	nameServer = "ns3.dnsimple.com"
)

var _ = XDescribe("type Provider", func() {
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
				GetRecords: func(ctx context.Context) ([]dns.RR, error) {
					var records []dns.RR
					return records, dnsimplex.Each(
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
							if rr, ok := convertRecord(rec); ok {
								records = append(records, rr)
							}
							return true, nil
						},
					)
				},
				DeleteRecords: func(ctx context.Context, service string) error {
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
							var err error
							if strings.Contains(rec.Name, service) {
								_, err = client.Zones.DeleteRecord(
									ctx,
									accountID,
									domain,
									rec.ID,
								)
							}
							return true, err
						},
					)
				},
			}
		},
	)
})

func convertRecord(rec dnsimple.ZoneRecord) (dns.RR, bool) {
	switch rec.Type {
	default:
		return nil, false
	case "PTR":
		return &dns.PTR{
			Hdr: dns.RR_Header{
				Name:   rec.Name + "." + domain + ".",
				Rrtype: dns.TypePTR,
				Class:  dns.ClassINET,
				Ttl:    uint32(rec.TTL),
			},
			Ptr: rec.Content + ".",
		}, true

	case "SRV":
		parts := strings.Split(rec.Content, " ")

		weight, err := strconv.Atoi(parts[0])
		if err != nil {
			panic(err)
		}

		port, err := strconv.Atoi(parts[1])
		if err != nil {
			panic(err)
		}

		return &dns.SRV{
			Hdr: dns.RR_Header{
				Name:   rec.Name + "." + domain + ".",
				Rrtype: dns.TypeSRV,
				Class:  dns.ClassINET,
				Ttl:    uint32(rec.TTL),
			},
			Target:   parts[2] + ".",
			Port:     uint16(port),
			Priority: uint16(rec.Priority),
			Weight:   uint16(weight),
		}, true

	case "TXT":
		return &dns.TXT{
			Hdr: dns.RR_Header{
				Name:   rec.Name + "." + domain + ".",
				Rrtype: dns.TypeTXT,
				Class:  dns.ClassINET,
				Ttl:    uint32(rec.TTL),
			},
			Txt: strings.Split(
				strings.Trim(rec.Content, `"`),
				`" "`,
			),
		}, true
	}
}
