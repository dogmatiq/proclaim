package provider

import (
	"github.com/dogmatiq/dissolve/dnssd"
)

// Advertiser is an interface for advertising DNS-SD service instances on a
// specific domain.
type Advertiser interface {
	dnssd.Advertiser

	// ID returns a data-structure that unique identifies this advertiser within
	// the provider that created it.
	ID() map[string]any
}
