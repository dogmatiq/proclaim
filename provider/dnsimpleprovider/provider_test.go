package dnsimpleprovider_test

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/dnsimple/dnsimple-go/dnsimple"
	"github.com/dogmatiq/dissolve/dnssd"
	"github.com/dogmatiq/proclaim/provider"
	. "github.com/dogmatiq/proclaim/provider/dnsimpleprovider"
	"github.com/dogmatiq/proclaim/provider/dnsimpleprovider/internal/dnsimplex"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// testDomain is the domain name used for testing in the DNSimple SANDBOX
// environment.
const (
	testAccountID = "1946"
	testDomain    = "dogmatiq.io"
)

var _ = Describe("type advertiser", func() {
	var (
		ctx    context.Context
		client *dnsimple.Client
		prov   *Provider
	)

	BeforeEach(func() {
		token := os.Getenv("DOGMATIQ_TEST_DNSIMPLE_TOKEN")
		if token == "" {
			Skip("DOGMATIQ_TEST_DNSIMPLE_TOKEN is not defined")
		}

		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 15*time.Second)
		DeferCleanup(cancel)

		client = dnsimple.NewClient(
			dnsimple.StaticTokenHTTPClient(
				ctx,
				token,
			),
		)
		client.BaseURL = "https://api.sandbox.dnsimple.com"

		prov = &Provider{
			API: client,
		}

		deleteAllRecords(client)
	})

	AfterEach(func() {
		deleteAllRecords(client)
	})

	When("the domain exists", func() {
		var (
			inst1, inst2 dnssd.ServiceInstance
			adv          provider.Advertiser
		)

		BeforeEach(func() {
			attr1 := dnssd.Attributes{}
			attr1.Set("key", []byte("value"))
			attr2 := dnssd.Attributes{}
			attr2.SetFlag("flag")

			inst1 = dnssd.ServiceInstance{
				Instance:    "instance-1",
				ServiceType: "_proclaim._udp",
				Domain:      testDomain,
				TargetHost:  "host1.example.com",
				TargetPort:  443,
				Priority:    10,
				Weight:      20,
				Attributes:  []dnssd.Attributes{attr1, attr2},
				TTL:         1 * time.Minute,
			}

			inst2 = dnssd.ServiceInstance{
				Instance:    "instance-2",
				ServiceType: "_proclaim._udp",
				Domain:      testDomain,
				TargetHost:  "host2.example.com",
				TargetPort:  8080,
				Priority:    100,
				Weight:      200,
				Attributes:  []dnssd.Attributes{attr1},
				TTL:         2 * time.Minute,
			}

			var ok bool
			var err error
			adv, ok, err = prov.AdvertiserByDomain(ctx, testDomain)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(ok).To(BeTrue())
		})

		It("can advertise an instance", func() {
			res, err := adv.Advertise(ctx, inst1)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res).To(Equal(provider.AdvertisedNewInstance))

			expectRecords(
				ctx,
				client,
				`_proclaim._udp.dogmatiq.io. 60 IN PTR instance-1._proclaim._udp.dogmatiq.io.`,
				`instance-1._proclaim._udp.dogmatiq.io. 60 IN SRV 10 20 443 host1.example.com`,
				`instance-1._proclaim._udp.dogmatiq.io. 60 IN TXT "\"key=value\""`,
				`instance-1._proclaim._udp.dogmatiq.io. 60 IN TXT "\"flag\""`,
			)
		})

		It("can advertise multiple instances of the same service type", func() {
			res, err := adv.Advertise(ctx, inst1)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res).To(Equal(provider.AdvertisedNewInstance))

			res, err = adv.Advertise(ctx, inst2)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res).To(Equal(provider.AdvertisedNewInstance))

			expectRecords(
				ctx,
				client,
				`_proclaim._udp.dogmatiq.io. 60 IN PTR instance-1._proclaim._udp.dogmatiq.io.`,
				`instance-1._proclaim._udp.dogmatiq.io. 60 IN SRV 10 20 443 host1.example.com`,
				`instance-1._proclaim._udp.dogmatiq.io. 60 IN TXT "\"key=value\""`,
				`instance-1._proclaim._udp.dogmatiq.io. 60 IN TXT "\"flag\""`,

				`_proclaim._udp.dogmatiq.io. 120 IN PTR instance-2._proclaim._udp.dogmatiq.io.`,
				`instance-2._proclaim._udp.dogmatiq.io. 120 IN SRV 100 200 8080 host2.example.com`,
				`instance-2._proclaim._udp.dogmatiq.io. 120 IN TXT "\"key=value\""`,
			)
		})

		It("can unadvertise the only instance of a service", func() {
			_, err := adv.Advertise(ctx, inst1)
			Expect(err).ShouldNot(HaveOccurred())

			res, err := adv.Unadvertise(ctx, inst1)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res).To(Equal(provider.UnadvertisedExistingInstance))

			expectRecords(
				ctx,
				client,
			)
		})

		It("can unadvertise the an instance", func() {
			_, err := adv.Advertise(ctx, inst1)
			Expect(err).ShouldNot(HaveOccurred())

			_, err = adv.Advertise(ctx, inst2)
			Expect(err).ShouldNot(HaveOccurred())

			res, err := adv.Unadvertise(ctx, inst1)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res).To(Equal(provider.UnadvertisedExistingInstance))

			expectRecords(
				ctx,
				client,
				`_proclaim._udp.dogmatiq.io. 120 IN PTR instance-2._proclaim._udp.dogmatiq.io.`,
				`instance-2._proclaim._udp.dogmatiq.io. 120 IN SRV 100 200 8080 host2.example.com`,
				`instance-2._proclaim._udp.dogmatiq.io. 120 IN TXT "\"key=value\""`,
			)
		})

		It("can update an existing instance", func() {
			res, err := adv.Advertise(ctx, inst1)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res).To(Equal(provider.AdvertisedNewInstance))

			inst1.TargetHost = "updated.example.com"
			inst1.TargetPort = 8443
			inst1.Priority = 1000
			inst1.Weight = 2000
			inst1.TTL = 45 * time.Second
			inst1.Attributes[0].Set("key", []byte("updated"))

			res, err = adv.Advertise(ctx, inst1)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res).To(Equal(provider.UpdatedExistingInstance))

			expectRecords(
				ctx,
				client,
				`_proclaim._udp.dogmatiq.io. 45 IN PTR instance-1._proclaim._udp.dogmatiq.io.`,
				`instance-1._proclaim._udp.dogmatiq.io. 45 IN SRV 1000 2000 8443 updated.example.com`,
				`instance-1._proclaim._udp.dogmatiq.io. 45 IN TXT "\"key=updated\""`,
				`instance-1._proclaim._udp.dogmatiq.io. 45 IN TXT "\"flag\""`,
			)
		})

		It("ignores an existing identical instance", func() {
			res, err := adv.Advertise(ctx, inst1)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res).To(Equal(provider.AdvertisedNewInstance))

			res, err = adv.Advertise(ctx, inst1)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res).To(Equal(provider.InstanceAlreadyAdvertised))

			expectRecords(
				ctx,
				client,
				`_proclaim._udp.dogmatiq.io. 60 IN PTR instance-1._proclaim._udp.dogmatiq.io.`,
				`instance-1._proclaim._udp.dogmatiq.io. 60 IN SRV 10 20 443 host1.example.com`,
				`instance-1._proclaim._udp.dogmatiq.io. 60 IN TXT "\"key=value\""`,
				`instance-1._proclaim._udp.dogmatiq.io. 60 IN TXT "\"flag\""`,
			)
		})
	})

	When("the domain does not exist", func() {
		Describe("func AdvertiserByDomain()", func() {
			It("returns false", func() {
				_, ok, err := prov.AdvertiserByDomain(ctx, "non-existent."+testDomain)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(ok).To(BeFalse())
			})
		})
	})
})

func deleteAllRecords(client *dnsimple.Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dnsimplex.Each(
		ctx,
		func(opts dnsimple.ListOptions) (*dnsimple.Pagination, []dnsimple.ZoneRecord, error) {
			res, err := client.Zones.ListRecords(
				ctx,
				testAccountID,
				testDomain,
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
					testAccountID,
					testDomain,
					rec.ID,
				)
				return true, err
			}
		},
	)
}

// expectRecords asserts that the test domain's zone file contains the expected
// lines. Note that there is currently a bug in DNSimple's zone export feature
// that causes it to double-escape TXT record content.
func expectRecords(ctx context.Context, client *dnsimple.Client, expect ...string) {
	res, err := client.Zones.GetZoneFile(ctx, testAccountID, testDomain)
	Expect(err).ShouldNot(HaveOccurred())

	lines := strings.Split(
		strings.TrimSpace(res.Data.Zone),
		"\n",
	)

loop:
	for len(lines) > 0 {
		line := lines[0]

		switch {
		case strings.HasPrefix(line, "$"):
		case strings.Contains(line, " IN SOA "):
		case strings.Contains(line, " IN NS "):
		default:
			break loop
		}

		lines = lines[1:]
	}

	Expect(lines).To(ConsistOf(expect))
}
