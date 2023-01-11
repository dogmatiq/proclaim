package driver

import (
	"context"

	"github.com/dogmatiq/dissolve/dnssd"
	"github.com/go-logr/logr"
)

// Driver is an interface for advertising DNS-SD service instances on domains
// hosted by a specific hosting provider.
type Driver interface {
	// Name returns a human-readable name for the driver.
	Name() string

	// AdvertiserForDomain returns the Advertiser used to advertise services on the
	// given domain.
	//
	// ok is false if this driver does not manage the given domain.
	AdvertiserForDomain(
		ctx context.Context,
		logger logr.Logger,
		domain string,
	) (_ Advertiser, ok bool, _ error)
}

// Advertiser is an interface for advertising DNS-SD service instances on a
// specific domain.
type Advertiser interface {
	// Advertise adds/updates DNS records to advertise the given service
	// instance.
	Advertise(
		ctx context.Context,
		logger logr.Logger,
		inst dnssd.ServiceInstance,
	) error

	// Advertise removes/updates DNS records to stop advertising the given
	// service instance.
	Unadvertise(
		ctx context.Context,
		logger logr.Logger,
		inst dnssd.ServiceInstance,
	) error
}
