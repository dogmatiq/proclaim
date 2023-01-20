package main

import (
	"os"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/dogmatiq/imbue"
	"github.com/dogmatiq/proclaim/provider/route53provider"
	"github.com/dogmatiq/proclaim/reconciler"
)

func init() {
	imbue.Decorate0(
		container,
		func(
			ctx imbue.Context,
			r *reconciler.Reconciler,
		) (*reconciler.Reconciler, error) {
			if os.Getenv("ROUTE53_ENABLED") == "" {
				return r, nil
			}

			s, err := session.NewSession()
			if err != nil {
				return nil, err
			}

			r.Providers = append(
				r.Providers,
				&route53provider.Provider{
					API: route53.New(s),
				},
			)

			return r, nil
		},
	)
}
