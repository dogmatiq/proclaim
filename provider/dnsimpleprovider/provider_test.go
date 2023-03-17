package dnsimpleprovider_test

import (
	"context"
	"os"
	"time"

	"github.com/dnsimple/dnsimple-go/dnsimple"
	"github.com/dogmatiq/dissolve/dnssd"
	"github.com/dogmatiq/proclaim/provider"
	. "github.com/dogmatiq/proclaim/provider/dnsimpleprovider"
	"github.com/dogmatiq/proclaim/provider/dnsimpleprovider/internal/dnsimplex"
	"github.com/miekg/dns"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/exp/slices"
)

// testDomain is the domain name used for testing in the DNSimple SANDBOX
// environment.
const (
	testService = "_test._udp"
	testDomain  = "proclaim-test.dogmatiq.io"
)

var _ = Describe("type advertiser", func() {
	var (
		ctx      context.Context
		resolver dnssd.UnicastResolver
		client   *dnsimple.Client
		prov     *Provider
	)

	expectInstanceToExist := func(expect dnssd.ServiceInstance) {
		for {
			actual, ok, err := resolver.LookupInstance(ctx, expect.Instance, expect.ServiceType, expect.Domain)
			Expect(err).ShouldNot(HaveOccurred())

			if ok {
				Expect(actual).To(Equal(expect))
				return
			}

			time.Sleep(50 * time.Millisecond)
		}
	}

	expectInstanceNotToExist := func(expect dnssd.ServiceInstance) {
		for {
			_, ok, err := resolver.LookupInstance(ctx, expect.Instance, expect.ServiceType, expect.Domain)
			Expect(err).ShouldNot(HaveOccurred())

			if !ok {
				return
			}

			time.Sleep(50 * time.Millisecond)
		}
	}

	expectEnumerableInstances := func(expect ...dnssd.ServiceInstance) {
		var names []string
		for _, inst := range expect {
			names = append(names, inst.Instance)
		}

		slices.Sort(names)

		for {
			instances, err := resolver.EnumerateInstances(ctx, testService, testDomain)
			Expect(err).ShouldNot(HaveOccurred())

			slices.Sort(instances)

			if slices.Equal(instances, names) {
				return
			}

			time.Sleep(50 * time.Millisecond)
		}
	}

	BeforeEach(func() {
		token := os.Getenv("DOGMATIQ_TEST_DNSIMPLE_TOKEN")
		if token == "" {
			Skip("DOGMATIQ_TEST_DNSIMPLE_TOKEN is not defined")
		}

		resolver = dnssd.UnicastResolver{
			Config: &dns.ClientConfig{
				Port:     "53",
				Ndots:    1,
				Timeout:  5,
				Attempts: 10,
				Servers: []string{
					"ns3.dnsimple.com",
				},
			},
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

		prov = &Provider{
			API: client,
		}

		deleteAllRecords(client)
		expectEnumerableInstances()
	})

	When("the provider can not advertise on the domain", func() {
		Describe("func AdvertiserByDomain()", func() {
			It("returns false", func() {
				_, ok, err := prov.AdvertiserByDomain(ctx, "non-existent."+testDomain)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(ok).To(BeFalse())
			})
		})
	})

	When("the provider can advertise on the domain", func() {
		var advertiser provider.Advertiser

		BeforeEach(func() {
			var (
				ok  bool
				err error
			)
			advertiser, ok, err = prov.AdvertiserByDomain(ctx, testDomain)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(ok).To(BeTrue())

			deleteAllRecords(client)
		})

		It("can advertise and unadvertise a single instance", func() {
			expect := dnssd.ServiceInstance{
				Instance:    "instance",
				ServiceType: testService,
				Domain:      testDomain,
				TargetHost:  "host.example.com",
				TargetPort:  443,
				Priority:    10,
				Weight:      20,
				TTL:         5 * time.Second,
			}

			expectInstanceNotToExist(expect)

			a, err := advertiser.Advertise(ctx, expect)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(a).To(Equal(provider.AdvertisedNewInstance))

			expectInstanceToExist(expect)
			expectEnumerableInstances(expect)

			u, err := advertiser.Unadvertise(ctx, expect)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(u).To(Equal(provider.UnadvertisedExistingInstance))

			expectInstanceNotToExist(expect)
			expectEnumerableInstances()
		})

		It("can advertise multiple instances of the same service type", func() {
			expect := []dnssd.ServiceInstance{
				{
					Instance:    "instance-1",
					ServiceType: testService,
					Domain:      testDomain,
					TargetHost:  "host1.example.com",
					TargetPort:  1000,
					Priority:    100,
					Weight:      10,
					TTL:         1 * time.Second,
				},
				{
					Instance:    "instance-2",
					ServiceType: testService,
					Domain:      testDomain,
					TargetHost:  "host2.example.com",
					TargetPort:  2000,
					Priority:    200,
					Weight:      20,
					TTL:         2 * time.Second,
				},
			}

			for _, inst := range expect {
				expectInstanceNotToExist(inst)

				res, err := advertiser.Advertise(ctx, inst)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res).To(Equal(provider.AdvertisedNewInstance))
			}

			// Check that all instances exist AFTER all the advertise calls.
			for _, inst := range expect {
				expectInstanceToExist(inst)
			}

			expectEnumerableInstances(expect...)
		})

		It("can update an existing instance", func() {
			expect := dnssd.ServiceInstance{
				Instance:    "instance",
				ServiceType: testService,
				Domain:      testDomain,
				TargetHost:  "host.example.com",
				TargetPort:  443,
				Priority:    10,
				Weight:      20,
				TTL:         5 * time.Second,
				Attributes: []dnssd.Attributes{
					*dnssd.
						NewAttributes().
						Set("key", []byte("value")),
				},
			}

			expectInstanceNotToExist(expect)

			_, err := advertiser.Advertise(ctx, expect)
			Expect(err).ShouldNot(HaveOccurred())

			expectInstanceToExist(expect)

			expect.TargetHost = "updated.example.com"
			expect.TargetPort = 8443
			expect.Priority = 1000
			expect.Weight = 2000
			expect.TTL = 45 * time.Second
			expect.Attributes[0].Set("key", []byte("updated"))

			res, err := advertiser.Advertise(ctx, expect)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res).To(Equal(provider.UpdatedExistingInstance))

			expectInstanceToExist(expect)
		})

		It("ignores an existing identical instance", func() {
			expect := dnssd.ServiceInstance{
				Instance:    "instance",
				ServiceType: testService,
				Domain:      testDomain,
				TargetHost:  "host.example.com",
				TargetPort:  443,
				Priority:    10,
				Weight:      20,
				TTL:         5 * time.Second,
				Attributes: []dnssd.Attributes{
					*dnssd.
						NewAttributes().
						Set("key", []byte("value")),
				},
			}

			expectInstanceNotToExist(expect)

			_, err := advertiser.Advertise(ctx, expect)
			Expect(err).ShouldNot(HaveOccurred())

			res, err := advertiser.Advertise(ctx, expect)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res).To(Equal(provider.InstanceAlreadyAdvertised))

			expectInstanceToExist(expect)
		})
	})

})

func deleteAllRecords(client *dnsimple.Client) {
	const testAccountID = "12018"

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := dnsimplex.Each(
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
	Expect(err).ShouldNot(HaveOccurred())
}
