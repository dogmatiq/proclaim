package dnsimpleprovider_test

import (
	"context"
	"os"
	"testing"

	"github.com/dnsimple/dnsimple-go/v4/dnsimple"
	. "github.com/dogmatiq/proclaim/provider/dnsimpleprovider"
	"github.com/dogmatiq/proclaim/provider/internal/providertest"
)

func TestProvider(t *testing.T) {
	token := os.Getenv("DOGMATIQ_TEST_DNSIMPLE_TOKEN")
	if token == "" {
		t.Skip("DOGMATIQ_TEST_DNSIMPLE_TOKEN is not defined")
	}

	client := dnsimple.NewClient(
		dnsimple.StaticTokenHTTPClient(
			context.Background(),
			token,
		),
	)

	providertest.Run(
		t,
		providertest.TestContext{
			Provider: &Provider{
				Client: client,
			},
			Domain: "dissolve-test.dogmatiq.io",
		},
	)
}
