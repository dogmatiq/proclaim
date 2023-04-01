package main

import (
	"context"
	"log"

	"github.com/dogmatiq/ferrite"
	"github.com/dogmatiq/imbue"
	"github.com/dogmatiq/proclaim/crd"
	"github.com/dogmatiq/proclaim/provider"
	"github.com/dogmatiq/proclaim/reconciler"
	"github.com/go-logr/logr"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	controller "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

var container = imbue.New()

func main() {
	defer container.Close()
	ferrite.Init()

	ctx := controller.SetupSignalHandler()
	g := container.WaitGroup(ctx)

	imbue.Go2(
		g,
		func(
			ctx context.Context,
			providers []provider.Provider,
			l imbue.ByName[systemLogger, logr.Logger],
		) error {
			for _, p := range providers {
				l.Value().Info(
					"provider enabled",
					"id", p.ID(),
				)
			}
			return nil
		},
	)

	imbue.Go3(
		g,
		func(
			ctx context.Context,
			m manager.Manager,
			r *reconciler.InstanceReconciler,
			l imbue.ByName[systemLogger, logr.Logger],
		) error {
			err := builder.
				ControllerManagedBy(m).
				For(&crd.DNSSDServiceInstance{}).
				WithEventFilter(predicate.GenerationChangedPredicate{}).
				Complete(r)
			if err != nil {
				return err
			}

			return m.Start(ctx)
		},
	)

	// imbue.Go3(
	// 	g,
	// 	func(
	// 		ctx context.Context,
	// 		m manager.Manager,
	// 		r *subtype.Reconciler,
	// 		l imbue.ByName[systemLogger, logr.Logger],
	// 	) error {
	// 		err := builder.
	// 			ControllerManagedBy(m).
	// 			For(&crd.DNSSDServiceInstanceSubType{}).
	// 			WithEventFilter(predicate.GenerationChangedPredicate{}).
	// 			Complete(r)
	// 		if err != nil {
	// 			return err
	// 		}

	// 		return m.Start(ctx)
	// 	},
	// )

	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}
}
