package route53driver

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/route53/route53iface"
	"github.com/dogmatiq/proclaim"
	"github.com/go-logr/logr"
)

// Driver is an implementation of proclaim.Driver that advertises DNS-SD
// services on domains hosted by Amazon Route 53.
type Driver struct {
	API route53iface.Route53API
}

// Name returns a human-readable name for the driver.
func (d *Driver) Name() string {
	return "Amazon Route 53"
}

// AdvertiserForDomain returns the Advertiser used to advertise services on the
// given domain.
//
// ok is false if this driver does not manage the given domain.
func (d *Driver) AdvertiserForDomain(
	ctx context.Context,
	logger logr.Logger,
	domain string,
) (_ proclaim.Advertiser, ok bool, _ error) {
	domain += "."

	res, err := d.API.ListHostedZonesByNameWithContext(
		ctx,
		&route53.ListHostedZonesByNameInput{
			DNSName:  aws.String(domain),
			MaxItems: aws.String("1"),
		},
	)
	if err != nil {
		return nil, false, fmt.Errorf("unable to list hosted zones: %w", err)
	}

	if len(res.HostedZones) == 0 {
		return nil, false, nil
	}

	zone := res.HostedZones[0]

	if *zone.Name != domain {
		return nil, false, nil
	}

	return &advertiser{
		API:    d.API,
		ZoneID: zone.Id,
	}, true, nil
}
