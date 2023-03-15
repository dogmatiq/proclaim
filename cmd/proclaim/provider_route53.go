package main

import (
	"github.com/dogmatiq/ferrite"
	"github.com/dogmatiq/imbue"
	"github.com/dogmatiq/proclaim/reconciler"
)

var route53Enabled = ferrite.
	Bool("ROUTE53_ENABLED", "enable the AWS Route 53 provider").
	WithDefault(false).
	Required()

func init() {
	imbue.Decorate0(
		container,
		func(
			ctx imbue.Context,
			r *reconciler.Reconciler,
		) (*reconciler.Reconciler, error) {
			if !route53Enabled.Value() {
				return r, nil
			}

			// 		s, err := session.NewSession()
			// 		if err != nil {
			// 			return nil, err
			// 		}

			// 		r.Providers = append(
			// 			r.Providers,
			// 			&route53provider.Provider{
			// 				API: route53.New(s),
			// 			},
			// 		)

			return r, nil
		},
	)
}
