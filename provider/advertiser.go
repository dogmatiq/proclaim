package provider

import (
	"context"

	"github.com/dogmatiq/dissolve/dnssd"
)

// Advertiser is an interface for advertising DNS-SD service instances on a
// specific domain.
type Advertiser interface {
	// ID returns a data-structure that unique identifies this advertiser within
	// the provider that created it.
	ID() map[string]any

	// AdvertiseInstance adds/updates DNS records to advertise the given service
	// instance.
	AdvertiseInstance(ctx context.Context, inst dnssd.ServiceInstance) (ChangeSet, error)

	// Advertise removes/updates DNS records to stop advertising the given
	// service instance.
	UnadvertiseInstance(ctx context.Context, inst dnssd.ServiceInstanceName) (ChangeSet, error)
}

// ChangeSet describes the changes made to DNS records.
type ChangeSet struct {
	PTR Change
	SRV Change
	TXT Change
}

// Change is a bit-field that describes the changes made to a specific DNS
// record.
type Change int

const (
	// NoChange indicates that no changes were made.
	NoChange Change = 0

	// Created indicates that a record of this type was created.
	Created Change = 1 << iota

	// Updated indicates that a record of this type was updated.
	Updated

	// Deleted indicates that a record of this type was deleted.
	Deleted
)

// IsEmpty returns true if no changes were made.
func (cs ChangeSet) IsEmpty() bool {
	if cs.PTR != NoChange {
		return false
	}

	if cs.SRV != NoChange {
		return false
	}

	if cs.TXT != NoChange {
		return false
	}

	return true
}

// IsCreate returns true if the change set represents an entirely new
// instance.
func (cs ChangeSet) IsCreate() bool {
	return cs.PTR == Created || cs.SRV == Created
}
