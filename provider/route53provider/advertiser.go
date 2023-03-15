package route53provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/dogmatiq/dissolve/dnssd"
	"github.com/dogmatiq/proclaim/provider"
	"github.com/miekg/dns"
)

type advertiser struct {
	PartitionID string
	Client      *route53.Client
	ZoneID      string
}

func (a *advertiser) ID() string {
	return arn.ARN{
		Partition: a.PartitionID,
		Service:   "route53",
		Resource:  fmt.Sprintf("hostedzone/%s", a.ZoneID),
	}.String()
}

func (a *advertiser) Advertise(
	ctx context.Context,
	inst dnssd.ServiceInstance,
) (provider.AdvertiseResult, error) {
	cs := &types.ChangeBatch{
		Comment: aws.String(fmt.Sprintf(
			"dogmatiq/proclaim: advertising %s instance: %s ",
			inst.ServiceType,
			inst.Instance,
		)),
	}

	if err := a.syncPTR(ctx, inst, cs); err != nil {
		return provider.AdvertiseError, err
	}

	if err := a.syncSRV(ctx, inst, cs); err != nil {
		return provider.AdvertiseError, err
	}

	if err := a.syncTXT(ctx, inst, cs); err != nil {
		return provider.AdvertiseError, err
	}

	if len(cs.Changes) == 0 {
		return provider.InstanceAlreadyAdvertised, nil
	}

	if _, err := a.Client.ChangeResourceRecordSets(
		ctx,
		&route53.ChangeResourceRecordSetsInput{
			HostedZoneId: aws.String(a.ZoneID),
			ChangeBatch:  cs,
		},
	); err != nil {
		return provider.AdvertiseError, err
	}

	// If we had to create any of the resource record sets then the service was
	// never "fully" advertised, as all of the records (PTR, SRV and TXT) are
	// mandatory according to the DNS-SD spec.
	for _, c := range cs.Changes {
		if c.Action == types.ChangeActionCreate {
			return provider.AdvertisedNewInstance, nil
		}
	}

	return provider.UpdatedExistingInstance, nil
}

func (a *advertiser) Unadvertise(
	ctx context.Context,
	inst dnssd.ServiceInstance,
) (provider.UnadvertiseResult, error) {
	cs := &types.ChangeBatch{
		Comment: aws.String(fmt.Sprintf(
			"dogmatiq/proclaim: unadvertising %s instance: %s ",
			inst.ServiceType,
			inst.Instance,
		)),
	}

	if err := a.deletePTR(ctx, inst, cs); err != nil {
		return provider.UnadvertiseError, err
	}

	if err := a.deleteSRV(ctx, inst, cs); err != nil {
		return provider.UnadvertiseError, err
	}

	if err := a.deleteTXT(ctx, inst, cs); err != nil {
		return provider.UnadvertiseError, err
	}

	if len(cs.Changes) == 0 {
		return provider.InstanceNotAdvertised, nil
	}

	if _, err := a.Client.ChangeResourceRecordSets(
		ctx,
		&route53.ChangeResourceRecordSetsInput{
			HostedZoneId: aws.String(a.ZoneID),
			ChangeBatch:  cs,
		},
	); err != nil {
		return provider.UnadvertiseError, err
	}

	return provider.UnadvertisedExistingInstance, nil
}

func (a *advertiser) findResourceRecordSet(
	ctx context.Context,
	name *string,
	recordType types.RRType,
) (types.ResourceRecordSet, bool, error) {
	out, err := a.Client.ListResourceRecordSets(
		ctx,
		&route53.ListResourceRecordSetsInput{
			HostedZoneId:    aws.String(a.ZoneID),
			StartRecordName: name,
			StartRecordType: recordType,
			MaxItems:        aws.Int32(1),
		},
	)
	if err != nil {
		return types.ResourceRecordSet{}, false, err
	}

	if len(out.ResourceRecordSets) == 0 {
		return types.ResourceRecordSet{}, false, nil
	}

	set := out.ResourceRecordSets[0]

	if !strings.EqualFold(*set.Name, *name) {
		return types.ResourceRecordSet{}, false, nil
	}

	if set.Type != recordType {
		return types.ResourceRecordSet{}, false, nil
	}

	return set, true, nil
}

func instanceName(inst dnssd.ServiceInstance) *string {
	return aws.String(
		dnssd.ServiceInstanceName(inst.Instance, inst.ServiceType, inst.Domain) + ".",
	)
}

func serviceName(inst dnssd.ServiceInstance) *string {
	return aws.String(
		dnssd.InstanceEnumerationDomain(inst.ServiceType, inst.Domain) + ".",
	)
}

func convertRecords[
	R interface {
		Header() *dns.RR_Header
		String() string
	},
](records ...R) []types.ResourceRecord {
	var result []types.ResourceRecord

	for _, rec := range records {
		result = append(
			result,
			types.ResourceRecord{
				Value: aws.String(
					strings.TrimPrefix(
						rec.String(),
						rec.Header().String(),
					),
				),
			},
		)
	}

	return result
}
