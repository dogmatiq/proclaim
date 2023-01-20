package route53provider

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/dogmatiq/proclaim/provider"
)

// Provider is an implementation of provider.Provider that advertises DNS-SD
// services on domains hosted by Amazon Route 53.
type Provider struct {
	API *route53.Route53
}

// ID returns a short unique identifier for the provider.
func (d *Provider) ID() string {
	return "route53"
}

// AdvertiserByID returns the Advertiser with the given ID.
func (d *Provider) AdvertiserByID(ctx context.Context, id string) (provider.Advertiser, error) {
	panic("not implemented")
}

// AdvertiserByDomain returns the Advertiser used to advertise services on the
// given domain.
//
// ok is false if this provider does not manage the given domain.
func (d *Provider) AdvertiserByDomain(ctx context.Context, domain string) (_ provider.Advertiser, ok bool, _ error) {
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

	advertiserID := arn.ARN{
		Partition: d.API.ClientInfo.PartitionID,
		Service:   route53.ServiceName,
		Resource:  fmt.Sprintf("hostedzone/%s", *zone.Id),
	}.String()

	return &advertiser{
		API:          d.API,
		AdvertiserID: advertiserID,
		ZoneID:       zone.Id,
	}, true, nil
}
