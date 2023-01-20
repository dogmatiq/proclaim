package main

import (
	"errors"
	"os"

	"github.com/dnsimple/dnsimple-go/dnsimple"
	"github.com/dogmatiq/imbue"
	"github.com/dogmatiq/proclaim/provider/dnsimpleprovider"
	"github.com/dogmatiq/proclaim/reconciler"
)

func init() {
	imbue.Decorate0(
		container,
		func(
			ctx imbue.Context,
			r *reconciler.Reconciler,
		) (*reconciler.Reconciler, error) {
			if os.Getenv("DNSIMPLE_ENABLED") == "" {
				return r, nil
			}

			token := os.Getenv("DNSIMPLE_TOKEN")
			if token == "" {
				return nil, errors.New("DNSIMPLE_TOKEN must be set")
			}

			client := dnsimple.NewClient(
				dnsimple.StaticTokenHTTPClient(ctx, token),
			)

			if u := os.Getenv("DNSIMPLE_API_URL"); u != "" {
				client.BaseURL = u
			}

			r.Providers = append(
				r.Providers,
				&dnsimpleprovider.Provider{
					API: client,
				},
			)

			return r, nil
		},
	)
}
