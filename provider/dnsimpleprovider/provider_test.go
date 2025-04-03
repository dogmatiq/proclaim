package dnsimpleprovider_test

import (
	"context"
	"os"

	"github.com/dnsimple/dnsimple-go/v4/dnsimple"
	. "github.com/dogmatiq/proclaim/provider/dnsimpleprovider"
	"github.com/dogmatiq/proclaim/provider/internal/providertest"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
)

const (
	accountID = "12018"
	domain    = "dissolve-test.dogmatiq.io"
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
			}
		},
	)
})
