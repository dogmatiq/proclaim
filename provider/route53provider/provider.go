package route53provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	route53advertiser "github.com/dogmatiq/dissolve/dnssd/advertiser/route53"
	"github.com/dogmatiq/proclaim/provider"
)

const defaultPartition = "aws"

// Provider is an implementation of provider.Provider that advertises DNS-SD
// services on domains hosted by Amazon Route 53.
type Provider struct {
	Client      *route53.Client
	PartitionID string
}

// ID returns a short unique identifier for the provider.
func (p *Provider) ID() string {
	if id := p.partitionID(); id != defaultPartition {
		return fmt.Sprintf("route53-%s", id)
	}
	return "route53"
}

// Describe returns a human-readable description of the provider.
func (p *Provider) Describe() string {
	if id := p.partitionID(); id != defaultPartition {
		return fmt.Sprintf("Amazon Route 53 (%s)", id)
	}
	return "Amazon Route 53"
}

// AdvertiserByID returns the Advertiser with the given ID.
func (p *Provider) AdvertiserByID(
	ctx context.Context,
	id map[string]any,
) (provider.Advertiser, error) {
	zoneID, err := unmarshalAdvertiserID(id)
	if err != nil {
		return nil, err
	}

	if _, err := p.Client.GetHostedZone(
		ctx,
		&route53.GetHostedZoneInput{
			Id: aws.String(zoneID),
		},
	); err != nil {
		return nil, fmt.Errorf("unable to get hosted zone: %w", err)
	}

	return &advertiser{
		&route53advertiser.Advertiser{
			Client: p.Client,
		},
		zoneID,
	}, nil
}

// AdvertiserByDomain returns the Advertiser used to advertise services on
// the given domain.
//
// ok is false if this provider does not manage the given domain.
func (p *Provider) AdvertiserByDomain(
	ctx context.Context,
	domain string,
) (provider.Advertiser, bool, error) {
	domain += "."

	out, err := p.Client.ListHostedZonesByName(
		ctx,
		&route53.ListHostedZonesByNameInput{
			DNSName:  aws.String(domain),
			MaxItems: aws.Int32(1),
		},
	)
	if err != nil {
		return nil, false, fmt.Errorf("unable to list hosted zones: %w", err)
	}

	if len(out.HostedZones) == 0 {
		return nil, false, nil
	}

	zone := out.HostedZones[0]

	if *zone.Name != domain {
		return nil, false, nil
	}

	return &advertiser{
		&route53advertiser.Advertiser{
			Client: p.Client,
		},
		*zone.Id,
	}, true, nil
}

func (p *Provider) partitionID() string {
	if p.PartitionID == "" {
		return defaultPartition
	}
	return p.PartitionID
}

// marshalAdvertiserID returns the ID of the advertiser for the given zone.
func marshalAdvertiserID(zoneID string) map[string]any {
	return map[string]any{
		"hostedZoneID": zoneID,
	}
}

// unmarshalAdvertiserID parses an advertiser ID into its constituent parts.
func unmarshalAdvertiserID(id map[string]any) (zoneID string, err error) {
	zoneIDAny, ok := id["hostedZoneID"]
	if !ok {
		return "", errors.New("invalid advertiser ID: missing hostedZoneID key")
	}

	zoneID, ok = zoneIDAny.(string)
	if !ok || zoneID == "" {
		return "", errors.New("invalid advertiser ID: hostedZoneID must be a non-empty string")
	}

	return zoneID, nil
}
