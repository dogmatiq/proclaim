package providertest

import (
	"context"
	"time"

	"github.com/dogmatiq/dissolve/dnssd"
	"github.com/dogmatiq/proclaim/provider"
	"github.com/miekg/dns"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

// TestContext contains provider-specific testing-related information.
type TestContext struct {
	Provider   provider.Provider
	Domain     string
	NameServer string
}

// DeclareTestSuite declares a Ginkgo test suite for a provider implementation.
func DeclareTestSuite(
	setUp func(context.Context) TestContext,
) {
	const service = "_test._udp"

	ginkgo.Describe("Provider", func() {
		var (
			ctx      context.Context
			tctx     TestContext
			resolver *dnssd.UnicastResolver
		)

		ginkgo.BeforeEach(func() {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(context.Background(), 15*time.Second)
			ginkgo.DeferCleanup(cancel)

			tctx = setUp(ctx)

			resolver = &dnssd.UnicastResolver{
				Config: &dns.ClientConfig{
					Port:     "53",
					Ndots:    1,
					Timeout:  5,
					Attempts: 10,
					Servers:  []string{tctx.NameServer},
				},
			}

			expectInstanceListToConverge(ctx, resolver, service, tctx.Domain)
		})

		ginkgo.When("the provider can not advertise on the domain", func() {
			ginkgo.Describe("func AdvertiserByDomain()", func() {
				ginkgo.It("returns false", func() {
					_, ok, err := tctx.Provider.AdvertiserByDomain(ctx, "non-existent."+tctx.Domain)
					gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
					gomega.Expect(ok).To(gomega.BeFalse())
				})
			})
		})

		ginkgo.When("the provider can advertise on the domain", func() {
			var advertiser provider.Advertiser

			ginkgo.BeforeEach(func() {
				var (
					ok  bool
					err error
				)
				advertiser, ok, err = tctx.Provider.AdvertiserByDomain(ctx, tctx.Domain)
				gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
				gomega.Expect(ok).To(gomega.BeTrue())
			})

			ginkgo.It("can advertise and unadvertise a single instance", func() {
				expect := dnssd.ServiceInstance{
					Instance:    "instance",
					ServiceType: service,
					Domain:      tctx.Domain,
					TargetHost:  "host.example.com",
					TargetPort:  443,
					Priority:    10,
					Weight:      20,
					TTL:         5 * time.Second,
				}

				expectInstanceNotToExist(ctx, resolver, expect)

				a, err := advertiser.Advertise(ctx, expect)
				gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
				gomega.Expect(a).To(gomega.Equal(provider.AdvertisedNewInstance))

				expectInstanceToExist(ctx, resolver, expect)
				expectInstanceListToConverge(ctx, resolver, service, tctx.Domain, expect)

				u, err := advertiser.Unadvertise(ctx, expect)
				gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
				gomega.Expect(u).To(gomega.Equal(provider.UnadvertisedExistingInstance))

				expectInstanceNotToExist(ctx, resolver, expect)
				expectInstanceListToConverge(ctx, resolver, service, tctx.Domain)
			})

			ginkgo.It("can advertise multiple instances of the same service type", func() {
				expect := []dnssd.ServiceInstance{
					{
						Instance:    "instance-1",
						ServiceType: service,
						Domain:      tctx.Domain,
						TargetHost:  "host1.example.com",
						TargetPort:  1000,
						Priority:    100,
						Weight:      10,
						TTL:         1 * time.Second,
					},
					{
						Instance:    "instance-2",
						ServiceType: service,
						Domain:      tctx.Domain,
						TargetHost:  "host2.example.com",
						TargetPort:  2000,
						Priority:    200,
						Weight:      20,
						TTL:         2 * time.Second,
					},
				}

				for _, inst := range expect {
					expectInstanceNotToExist(ctx, resolver, inst)

					res, err := advertiser.Advertise(ctx, inst)
					gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
					gomega.Expect(res).To(gomega.Equal(provider.AdvertisedNewInstance))
				}

				// Check that all instances exist AFTER all the advertise calls.
				for _, inst := range expect {
					expectInstanceToExist(ctx, resolver, inst)
				}

				expectInstanceListToConverge(ctx, resolver, service, tctx.Domain, expect...)
			})

			ginkgo.It("can update an existing instance", func() {
				before := dnssd.ServiceInstance{
					Instance:    "instance",
					ServiceType: service,
					Domain:      tctx.Domain,
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

				expectInstanceNotToExist(ctx, resolver, before)

				_, err := advertiser.Advertise(ctx, before)
				gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

				expectInstanceToExist(ctx, resolver, before)

				after := dnssd.ServiceInstance{
					Instance:    "instance",
					ServiceType: service,
					Domain:      tctx.Domain,
					TargetHost:  "updated.example.com",
					TargetPort:  444,
					Priority:    11,
					Weight:      21,
					TTL:         6 * time.Second,
					Attributes: []dnssd.Attributes{
						*dnssd.
							NewAttributes().
							Set("key", []byte("updated")),
					},
				}

				res, err := advertiser.Advertise(ctx, after)
				gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
				gomega.Expect(res).To(gomega.Equal(provider.UpdatedExistingInstance))

				expectInstanceToConverge(ctx, resolver, after)
			})

			ginkgo.It("ignores an existing identical instance", func() {
				expect := dnssd.ServiceInstance{
					Instance:    "instance",
					ServiceType: service,
					Domain:      tctx.Domain,
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

				expectInstanceNotToExist(ctx, resolver, expect)

				_, err := advertiser.Advertise(ctx, expect)
				gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

				res, err := advertiser.Advertise(ctx, expect)
				gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
				gomega.Expect(res).To(gomega.Equal(provider.InstanceAlreadyAdvertised))

				expectInstanceToExist(ctx, resolver, expect)
			})
		})
	})
}
