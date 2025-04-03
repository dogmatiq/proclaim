package dnsimpleprovider

import (
	"github.com/dnsimple/dnsimple-go/v4/dnsimple"
	dnsimpleadvertiser "github.com/dogmatiq/dissolve/dnssd/advertiser/dnsimple"
)

type advertiser struct {
	*dnsimpleadvertiser.Advertiser
	Zone *dnsimple.Zone
}

func (a *advertiser) ID() map[string]any {
	return marshalAdvertiserID(a.Zone)
}
