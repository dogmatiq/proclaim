package main

import (
	"github.com/dogmatiq/dissolve/dnssd"
	"github.com/dogmatiq/imbue"
	"github.com/miekg/dns"
)

func init() {
	imbue.With1(
		container,
		func(
			ctx imbue.Context,
			cfg *dns.ClientConfig,
		) (*dnssd.UnicastResolver, error) {
			return &dnssd.UnicastResolver{
				Config: cfg,
			}, nil
		},
	)

	imbue.With0(
		container,
		func(
			ctx imbue.Context,
		) (*dns.ClientConfig, error) {
			return dns.ClientConfigFromFile("/etc/resolv.conf")
		},
	)
}
