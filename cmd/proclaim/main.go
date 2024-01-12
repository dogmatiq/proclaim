package main

import (
	"context"
	"log"

	"github.com/dogmatiq/ferrite"
	"github.com/dogmatiq/imbue"
	"github.com/dogmatiq/proclaim/crd"
	"github.com/dogmatiq/proclaim/reconciler"
	"github.com/go-logr/logr"
	"github.com/miekg/dns"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	controller "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

var container = imbue.New()

func main() {
	ferrite.Init()

	ctx := controller.SetupSignalHandler()

	if err := imbue.Invoke4(
		ctx,
		container,
		func(
			ctx context.Context,
			m manager.Manager,
			r *reconciler.Reconciler,
			c *dns.ClientConfig,
			l imbue.ByName[systemLogger, logr.Logger],
		) error {
			l.Value().Info(
				"DNS configuration loaded",
				"servers", c.Servers,
				"port", c.Port,
				"timeout", c.Timeout,
			)

			err := builder.
				ControllerManagedBy(m).
				For(&crd.DNSSDServiceInstance{}).
				WithEventFilter(predicate.GenerationChangedPredicate{}).
				Complete(r)
			if err != nil {
				return err
			}

			for _, p := range r.Providers {
				l.Value().Info(
					"provider enabled",
					"id", p.ID(),
				)
			}

			return m.Start(ctx)
		},
	); err != nil {
		log.Fatal(err)
	}
}
