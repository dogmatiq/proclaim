package main

import (
	"github.com/dnsimple/dnsimple-go/dnsimple"
	"github.com/dogmatiq/ferrite"
	"github.com/dogmatiq/imbue"
	"github.com/dogmatiq/proclaim/provider"
	"github.com/dogmatiq/proclaim/provider/dnsimpleprovider"
	"github.com/go-logr/logr"
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
	imbue.Decorate1(
		container,
		func(
			ctx imbue.Context,
			providers []provider.Provider,
			l imbue.ByName[providerLogger, logr.Logger],
		) ([]provider.Provider, error) {
			if !dnsimpleEnabled.Value() {
				return providers, nil
			}

			client := dnsimple.NewClient(
				dnsimple.StaticTokenHTTPClient(
					ctx,
					dnsimpleToken.Value(),
				),
			)
			client.BaseURL = dnsimpleURL.Value().String()

			return append(
				providers,
				&dnsimpleprovider.Provider{
					Client: client,
					Logger: l.Value(),
				},
			), nil
		},
	)
}
