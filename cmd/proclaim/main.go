package main

import (
	"context"
	"log"

	"github.com/dogmatiq/imbue"
	"github.com/dogmatiq/proclaim/crd"
	"github.com/dogmatiq/proclaim/reconciler"
	"github.com/go-logr/logr"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	controller "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var container = imbue.New()

func main() {
	ctx := controller.SetupSignalHandler()

	if err := imbue.Invoke3(
		ctx,
		container,
		func(
			ctx context.Context,
			m manager.Manager,
			r *reconciler.Reconciler,
			l logr.Logger,
		) error {
			err := builder.
				ControllerManagedBy(m).
				For(&crd.DNSSDServiceInstance{}).
				Complete(r)
			if err != nil {
				return err
			}

			for _, p := range r.Providers {
				l.Info(
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
