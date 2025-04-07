package main

import (
	"github.com/dnsimple/dnsimple-go/v4/dnsimple"
	"github.com/dogmatiq/ferrite"
	"github.com/dogmatiq/imbue"
	"github.com/dogmatiq/proclaim/provider/dnsimpleprovider"
	"github.com/dogmatiq/proclaim/reconciler"
)

var dnsimpleEnabled = ferrite.
	Bool("DNSIMPLE_ENABLED", "enable the DNSimple provider").
	WithDefault(false).
	Required()

var dnsimpleToken = ferrite.
	String("DNSIMPLE_TOKEN", "enable the DNSimple provider").
	WithSensitiveContent().
	Required(ferrite.RelevantIf(dnsimpleEnabled))

var dnsimpleURL = ferrite.
	URL("DNSIMPLE_API_URL", "the URL of the DNSimple API").
	WithDefault("https://api.dnsimple.com").
	Required(ferrite.RelevantIf(dnsimpleEnabled))

func init() {
	imbue.Decorate0(
		container,
		func(
			ctx imbue.Context,
			r *reconciler.Reconciler,
		) (*reconciler.Reconciler, error) {
			if !dnsimpleEnabled.Value() {
				return r, nil
			}

			client := dnsimple.NewClient(
				dnsimple.StaticTokenHTTPClient(
					ctx,
					dnsimpleToken.Value(),
				),
			)
			client.BaseURL = dnsimpleURL.Value().String()

			r.Providers = append(
				r.Providers,
				&dnsimpleprovider.Provider{
					Client: client,
				},
			)

			return r, nil
		},
	)
}
