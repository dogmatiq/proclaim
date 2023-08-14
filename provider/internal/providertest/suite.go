package providertest

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/dogmatiq/dissolve/dnssd"
	"github.com/dogmatiq/proclaim/provider"
	"github.com/miekg/dns"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

const (
	testTimeout     = 1 * time.Minute
	convergeTimeout = 100 * time.Millisecond
)

// TestContext contains provider-specific testing-related information.
type TestContext struct {
	Provider      provider.Provider
	Domain        string
	GetRecords    func(ctx context.Context) ([]dns.RR, error)
	DeleteRecords func(ctx context.Context, service string) error
}

// DeclareTestSuite declares a Ginkgo test suite for a provider implementation.
func DeclareTestSuite(
	setUp func(context.Context) TestContext,
) {
	ginkgo.Describe("Provider", func() {
		var (
			ctx      context.Context
			tctx     TestContext
			server   *server
			resolver *dnssd.UnicastResolver
			service  string
		)

		ginkgo.BeforeEach(func() {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(context.Background(), 10*time.Minute)
			ginkgo.DeferCleanup(cancel)

			service = fmt.Sprintf(
				"_%d_%d._udp",
				os.Getpid(),
				time.Now().Unix(),
			)

			tctx = setUp(ctx)

			var err error
			server, resolver, err = startServer()
			gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
		})

		ginkgo.AfterEach(func() {
			server.Stop()

			err := tctx.DeleteRecords(ctx, service)
			gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
		})

		expectEnumerateToMatch := func(expect ...dnssd.ServiceInstance) {
			records, err := tctx.GetRecords(ctx)
			gomega.ExpectWithOffset(1, err).ShouldNot(gomega.HaveOccurred())
			server.SetRecords(records)

			var names []string
			for _, inst := range expect {
				names = append(names, inst.Name)
			}

			instances, err := resolver.EnumerateInstances(ctx, service, tctx.Domain)
			gomega.ExpectWithOffset(1, err).ShouldNot(gomega.HaveOccurred())
			gomega.ExpectWithOffset(1, instances).To(gomega.ConsistOf(names))
		}

		expectLookupToMatch := func(expect dnssd.ServiceInstance) {
			records, err := tctx.GetRecords(ctx)
			gomega.ExpectWithOffset(1, err).ShouldNot(gomega.HaveOccurred())
			server.SetRecords(records)

			actual, ok, err := resolver.LookupInstance(
				ctx,
				expect.Name,
				expect.ServiceType,
				expect.Domain,
			)
			gomega.ExpectWithOffset(1, err).ShouldNot(gomega.HaveOccurred())
			gomega.ExpectWithOffset(1, ok).To(gomega.BeTrue(), "instance not found")

			if !actual.Equal(expect) {
				gomega.ExpectWithOffset(1, actual).To(gomega.Equal(expect))
			}
		}

		expectLookupToFail := func(inst dnssd.ServiceInstance) {
			records, err := tctx.GetRecords(ctx)
			gomega.ExpectWithOffset(1, err).ShouldNot(gomega.HaveOccurred())
			server.SetRecords(records)

			_, ok, err := resolver.LookupInstance(
				ctx,
				inst.Name,
				inst.ServiceType,
				inst.Domain,
			)
			gomega.ExpectWithOffset(1, err).ShouldNot(gomega.HaveOccurred())
			gomega.ExpectWithOffset(1, ok).To(gomega.BeFalse(), "instance not found unexpectedly")
		}

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

			ginkgo.It("can advertise and unadvertise instances", func() {
				expect := []dnssd.ServiceInstance{
					{
						ServiceInstanceName: dnssd.ServiceInstanceName{
							Name:        "instance-1",
							ServiceType: service,
							Domain:      tctx.Domain,
						},
						TargetHost: "host1.example.com",
						TargetPort: 1000,
						Priority:   100,
						Weight:     10,
						TTL:        1 * time.Second,
					},
					{
						ServiceInstanceName: dnssd.ServiceInstanceName{
							Name:        "instance-2",
							ServiceType: service,
							Domain:      tctx.Domain,
						},
						TargetHost: "host2.example.com",
						TargetPort: 2000,
						Priority:   200,
						Weight:     20,
						TTL:        2 * time.Second,
					},
				}

				for i, inst := range expect {
					cs, err := advertiser.Advertise(ctx, inst)
					gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
					gomega.Expect(cs.IsCreate()).To(gomega.BeTrue())

					expectEnumerateToMatch(expect[:i+1]...)
					expectLookupToMatch(inst)
				}

				// Check that all instances still exist after they have all the
				// advertise calls.
				for _, inst := range expect {
					expectLookupToMatch(inst)
				}

				expectEnumerateToMatch(expect...)

				for i, inst := range expect {
					cs, err := advertiser.Unadvertise(ctx, inst)
					gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
					gomega.Expect(cs.IsEmpty()).To(gomega.BeFalse())

					expectEnumerateToMatch(expect[i+1:]...)
					expectLookupToFail(inst)
				}

				expectEnumerateToMatch()
			})

			ginkgo.It("can update an existing instance", func() {
				before := dnssd.ServiceInstance{
					ServiceInstanceName: dnssd.ServiceInstanceName{
						Name:        "instance",
						ServiceType: service,
						Domain:      tctx.Domain,
					},
					TargetHost: "host.example.com",
					TargetPort: 443,
					Priority:   10,
					Weight:     20,
					TTL:        5 * time.Second,
					Attributes: []dnssd.Attributes{
						dnssd.
							NewAttributes().
							WithPair("key", []byte("value")),
					},
				}

				_, err := advertiser.Advertise(ctx, before)
				gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

				after := dnssd.ServiceInstance{
					ServiceInstanceName: dnssd.ServiceInstanceName{
						Name:        "instance",
						ServiceType: service,
						Domain:      tctx.Domain,
					},
					TargetHost: "updated.example.com",
					TargetPort: 444,
					Priority:   11,
					Weight:     21,
					TTL:        6 * time.Second,
					Attributes: []dnssd.Attributes{
						dnssd.
							NewAttributes().
							WithPair("key", []byte("updated")),
					},
				}

				cs, err := advertiser.Advertise(ctx, after)
				gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
				gomega.Expect(cs.IsCreate()).To(gomega.BeFalse())
				gomega.Expect(cs.IsEmpty()).To(gomega.BeFalse())

				expectLookupToMatch(after)
			})

			ginkgo.It("ignores an existing identical instance", func() {
				expect := dnssd.ServiceInstance{
					ServiceInstanceName: dnssd.ServiceInstanceName{
						Name:        "instance",
						ServiceType: service,
						Domain:      tctx.Domain,
					},
					TargetHost: "host.example.com",
					TargetPort: 443,
					Priority:   10,
					Weight:     20,
					TTL:        5 * time.Second,
					Attributes: []dnssd.Attributes{
						dnssd.
							NewAttributes().
							WithPair("key", []byte("value")),
					},
				}

				_, err := advertiser.Advertise(ctx, expect)
				gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

				cs, err := advertiser.Advertise(ctx, expect)
				gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
				gomega.Expect(cs.IsEmpty()).To(gomega.BeTrue())

				expectLookupToMatch(expect)
			})

			ginkgo.It("does not fail when unadvertising a non-existent instance", func() {
				inst := dnssd.ServiceInstance{
					ServiceInstanceName: dnssd.ServiceInstanceName{
						Name:        "instance",
						ServiceType: service,
						Domain:      tctx.Domain,
					},
					TargetHost: "host.example.com",
					TargetPort: 443,
					Priority:   10,
					Weight:     20,
					TTL:        5 * time.Second,
					Attributes: []dnssd.Attributes{
						dnssd.
							NewAttributes().
							WithPair("key", []byte("value")),
					},
				}

				cs, err := advertiser.Unadvertise(ctx, inst)
				gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
				gomega.Expect(cs.IsEmpty()).To(gomega.BeTrue())
			})
		})
	})
}
