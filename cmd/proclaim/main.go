package main

import (
	"os"

	"github.com/dogmatiq/proclaim"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	runtime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	setupLog = runtime.Log.WithName("setup")
)

func main() {
	runtime.SetLogger(zap.New())

	m, err := runtime.NewManager(runtime.GetConfigOrDie(), runtime.Options{})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	err = proclaim.SchemeBuilder.AddToScheme(m.GetScheme())
	if err != nil {
		setupLog.Error(err, "unable to add scheme")
		os.Exit(1)
	}

	err = runtime.NewControllerManagedBy(m).
		For(&proclaim.Instance{}).
		Complete(&proclaim.Reconciler{
			Client: m.GetClient(),
			Scheme: m.GetScheme(),
		})
	if err != nil {
		setupLog.Error(err, "unable to create controller")
		os.Exit(1)
	}

	// err = runtime.NewWebhookManagedBy(m).
	// 	For(&proclaim.Instance{}).
	// 	Complete()
	// if err != nil {
	// 	setupLog.Error(err, "unable to create webhook")
	// 	os.Exit(1)
	// }

	setupLog.Info("starting manager")
	if err := m.Start(runtime.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
