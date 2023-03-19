package main

import (
	"os"

	"github.com/dogmatiq/imbue"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

type (
	systemLogger   imbue.Name[logr.Logger]
	providerLogger imbue.Name[logr.Logger]
)

func init() {
	imbue.With0Named[systemLogger](
		container,
		func(
			ctx imbue.Context,
		) (logr.Logger, error) {
			return zap.New(
				zap.UseDevMode(
					os.Getenv("DEBUG") != "",
				),
			), nil
		},
	)

	imbue.With1Named[providerLogger](
		container,
		func(
			ctx imbue.Context,
			l imbue.ByName[systemLogger, logr.Logger],
		) (logr.Logger, error) {
			return l.Value().V(1), nil
		},
	)
}
