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
	Advertise(ctx context.Context, inst dnssd.ServiceInstance) (AdvertiseResult, error)

	// Advertise removes/updates DNS records to stop advertising the given
	// service instance.
	Unadvertise(ctx context.Context, inst dnssd.ServiceInstance) (UnadvertiseResult, error)
}

// AdvertiseResult is an enumeration of the possible results of an Advertise()
// call.
type AdvertiseResult int

const (
	// AdvertiseError indicates that an error occurred while attempting to
	// advertise the service instance.
	AdvertiseError AdvertiseResult = iota

	// InstanceAlreadyAdvertised indicates that the service instance was already
	// advertised correctly, and no changes were made.
	InstanceAlreadyAdvertised

	// AdvertisedNewInstance indicates that the service instance was not previously
	// advertised and has been advertised successfully.
	AdvertisedNewInstance

	// UpdatedExistingInstance indicates that the service instance was previously advertised
	// but changes were necessary to at least one DNS record.
	UpdatedExistingInstance
)

// UnadvertiseResult is an enumeration of the possible results of an
// Unadvertise() call.
type UnadvertiseResult int

const (
	// UnadvertiseError indicates that an error occurred while attempting to
	// advertise the service instance.
	UnadvertiseError UnadvertiseResult = iota

	// InstanceNotAdvertised indicates that the service instance was not previously
	// advertised and no changes were made.
	InstanceNotAdvertised

	// UnadvertisedExistingInstance indicates that the service instance was previously
	// advertised and has been unadvertised successfully.
	UnadvertisedExistingInstance
)
