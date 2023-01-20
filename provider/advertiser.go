package provider

import (
	"context"

	"github.com/dogmatiq/dissolve/dnssd"
)

// Advertiser is an interface for advertising DNS-SD service instances on a
// specific domain.
type Advertiser interface {
	// ID returns a unique identifier for the advertiser.
	//
	// The identifier must uniquely describe this advertiser within the context
	// of the provider that created it.
	ID() string

	// Advertise adds/updates DNS records to advertise the given service
	// instance.
	Advertise(ctx context.Context, inst dnssd.ServiceInstance) error

	// Advertise removes/updates DNS records to stop advertising the given
	// service instance.
	Unadvertise(ctx context.Context, inst dnssd.ServiceInstance) error
}
