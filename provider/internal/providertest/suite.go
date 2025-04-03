package providertest

import (
	"context"
	"time"

	"github.com/dogmatiq/proclaim/provider"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

const testTimeout = 1 * time.Minute

// TestContext contains provider-specific testing-related information.
type TestContext struct {
	Provider provider.Provider
	Domain   string
}

// DeclareTestSuite declares a Ginkgo test suite for a provider implementation.
func DeclareTestSuite(
	setUp func(context.Context) TestContext,
) {
	ginkgo.Describe("Provider", func() {
		var (
			ctx  context.Context
			tctx TestContext
		)

		ginkgo.BeforeEach(func() {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(context.Background(), 10*time.Minute)
			ginkgo.DeferCleanup(cancel)

			tctx = setUp(ctx)
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
			ginkgo.Describe("func AdvertiserByDomain()", func() {
				ginkgo.It("returns an advertiser", func() {
					advertiser, ok, err := tctx.Provider.AdvertiserByDomain(ctx, tctx.Domain)
					gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
					gomega.Expect(ok).To(gomega.BeTrue())
					gomega.Expect(advertiser).NotTo(gomega.BeNil())
				})
			})
		})
	})
}
