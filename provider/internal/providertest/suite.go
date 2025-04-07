package providertest

import (
	"context"
	"testing"
	"time"

	"github.com/dogmatiq/proclaim/provider"
)

const testTimeout = 1 * time.Minute

// TestContext contains provider-specific testing-related information.
type TestContext struct {
	Provider provider.Provider
	Domain   string
}

// Run executes the provider test suite.
func Run(
	t *testing.T,
	tctx TestContext,
) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	t.Cleanup(cancel)

	t.Run("provider", func(t *testing.T) {
		t.Run("AdvertiserByDomain()", func(t *testing.T) {
			t.Run("when the provider can advertise on the domain", func(t *testing.T) {
				t.Run("it returns an advertiser", func(t *testing.T) {
					advertiser, ok, err := tctx.Provider.AdvertiserByDomain(ctx, tctx.Domain)
					if err != nil {
						t.Fatal(err)
					}

					if !ok {
						t.Fatal("expected ok to be true")
					}

					if advertiser == nil {
						t.Fatal("expected advertiser to be non-nil")
					}
				})
			})

			t.Run("when the provider can not advertise on the domain", func(t *testing.T) {
				t.Run("it returns false", func(t *testing.T) {
					_, ok, err := tctx.Provider.AdvertiserByDomain(ctx, "non-existent."+tctx.Domain)
					if err != nil {
						t.Fatal(err)
					}

					if ok {
						t.Fatal("expected ok to be false")
					}
				})
			})
		})

		t.Run("AdvertiserByID()", func(t *testing.T) {
			t.Run("it returns the advertiser", func(t *testing.T) {
				advertiser, ok, err := tctx.Provider.AdvertiserByDomain(ctx, tctx.Domain)
				if err != nil {
					t.Fatal(err)
				}
				if !ok {
					t.Fatal("could not find advertiser by domain")
				}

				advertiser, err = tctx.Provider.AdvertiserByID(ctx, advertiser.ID())
				if err != nil {
					t.Fatal(err)
				}
				if advertiser == nil {
					t.Fatal("expected advertiser to be non-nil")
				}
			})

			t.Run("returns an error when passed an invalid ID", func(t *testing.T) {
				_, err := tctx.Provider.AdvertiserByID(ctx, map[string]any{})
				if err == nil {
					t.Fatal("expected non-nil error")
				}
			})
		})
	})
}
