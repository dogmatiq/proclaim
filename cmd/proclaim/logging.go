package main

import (
	"os"

	"github.com/dogmatiq/imbue"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func init() {
	imbue.With0(
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
}
