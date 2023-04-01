package main

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/dogmatiq/ferrite"
	"github.com/dogmatiq/imbue"
	"github.com/dogmatiq/proclaim/provider"
	"github.com/dogmatiq/proclaim/provider/route53provider"
	"github.com/go-logr/logr"
)

var route53Enabled = ferrite.
	Bool("ROUTE53_ENABLED", "enable the AWS Route 53 provider").
	WithDefault(false).
	Required()

func init() {
	imbue.Decorate2(
		container,
		func(
			ctx imbue.Context,
			providers []provider.Provider,
			c imbue.Optional[*route53.Client],
			l imbue.ByName[providerLogger, logr.Logger],
		) ([]provider.Provider, error) {
			if !route53Enabled.Value() {
				return providers, nil
			}

			cli, err := c.Value()
			if err != nil {
				return nil, err
			}

			return append(
				providers,
				&route53provider.Provider{
					Client: cli,
					Logger: l.Value(),
				},
			), nil
		},
	)

	imbue.With1(
		container,
		func(
			ctx imbue.Context,
			cfg aws.Config,
		) (*route53.Client, error) {
			return route53.NewFromConfig(cfg), nil
		},
	)

	imbue.With0(
		container,
		func(
			ctx imbue.Context,
		) (aws.Config, error) {
			return config.LoadDefaultConfig(ctx)
		},
	)
}
