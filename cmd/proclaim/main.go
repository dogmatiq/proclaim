package main

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/dogmatiq/proclaim"
	"github.com/dogmatiq/proclaim/driver/route53driver"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	runtime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func main() {
	m, err := runtime.NewManager(
		runtime.GetConfigOrDie(),
		runtime.Options{
			Logger: zap.New(zap.UseDevMode(true)),
		},
	)
	if err != nil {
		panic(err)
	}

	if err := proclaim.SchemeBuilder.AddToScheme(m.GetScheme()); err != nil {
		panic(err)
	}

	if err := runtime.
		NewControllerManagedBy(m).
		For(&proclaim.DNSSDServiceInstance{}).
		Complete(&proclaim.Reconciler{
			Client: m.GetClient(),
			Drivers: []proclaim.Driver{
				&route53driver.Driver{
					API: route53.New(
						session.Must(session.NewSession()),
					),
				},
			},
		}); err != nil {
		panic(err)
	}

	ctx := runtime.SetupSignalHandler()

	if err := m.Start(ctx); err != nil {
		panic(err)
	}
}
