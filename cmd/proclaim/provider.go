package main

import (
	"github.com/dogmatiq/imbue"
	"github.com/dogmatiq/proclaim/provider"
)

func init() {
	imbue.With0(
		container,
		func(
			ctx imbue.Context,
		) ([]provider.Provider, error) {
			return nil, nil
		},
	)
}
