package provider

import (
	"context"
	"time"
)

// Timeout is the timeout for all provider operations.
const Timeout = 10 * time.Second

// Provider is an interface for advertising DNS-SD service instances on domains
// hosted by a specific hosting provider.
type Provider interface {
	// ID returns a unique identifier for the provider.
	ID() string

	// Describe returns a human-readable description of the provider.
	Describe() string

	// AdvertiserByID returns the Advertiser with the given ID.
	AdvertiserByID(ctx context.Context, id string) (Advertiser, error)

	// AdvertiserByDomain returns the Advertiser used to advertise services on
	// the given domain.
	//
	// ok is false if this provider does not manage the given domain.
	AdvertiserByDomain(ctx context.Context, domain string) (_ Advertiser, ok bool, _ error)
}
