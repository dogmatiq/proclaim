package route53provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/dogmatiq/dissolve/dnssd"
	"github.com/dogmatiq/proclaim/provider"
	"github.com/go-logr/logr"
	"github.com/miekg/dns"
)

type advertiser struct {
	Client *route53.Client
	ZoneID string
	Logger logr.Logger
}

func (a *advertiser) ID() string {
	return a.ZoneID
}

func (a *advertiser) Advertise(
	ctx context.Context,
	inst dnssd.ServiceInstance,
) (provider.ChangeSet, error) {
	cs := &types.ChangeBatch{
		Comment: aws.String(fmt.Sprintf(
			"dogmatiq/proclaim: advertising %s instance: %s ",
			inst.ServiceType,
			inst.Name,
		)),
	}

	if err := a.syncPTR(ctx, inst, cs); err != nil {
		return provider.ChangeSet{}, err
	}

	if err := a.syncSRV(ctx, inst, cs); err != nil {
		return provider.ChangeSet{}, err
	}

	if err := a.syncTXT(ctx, inst, cs); err != nil {
		return provider.ChangeSet{}, err
	}

	return a.apply(ctx, cs)
}

func (a *advertiser) Unadvertise(
	ctx context.Context,
	inst dnssd.ServiceInstance,
) (provider.ChangeSet, error) {
	cs := &types.ChangeBatch{
		Comment: aws.String(fmt.Sprintf(
			"dogmatiq/proclaim: unadvertising %s instance: %s ",
			inst.ServiceType,
			inst.Name,
		)),
	}

	if err := a.deletePTR(ctx, inst, cs); err != nil {
		return provider.ChangeSet{}, err
	}

	if err := a.deleteSRV(ctx, inst, cs); err != nil {
		return provider.ChangeSet{}, err
	}

	if err := a.deleteTXT(ctx, inst, cs); err != nil {
		return provider.ChangeSet{}, err
	}

	return a.apply(ctx, cs)
}

func (a *advertiser) apply(
	ctx context.Context,
	cs *types.ChangeBatch,
) (provider.ChangeSet, error) {
	if len(cs.Changes) == 0 {
		return provider.ChangeSet{}, nil
	}

	_, err := a.Client.ChangeResourceRecordSets(
		ctx,
		&route53.ChangeResourceRecordSetsInput{
			HostedZoneId: aws.String(a.ZoneID),
			ChangeBatch:  cs,
		},
	)
	if err != nil {
		return provider.ChangeSet{}, err
	}

	var result provider.ChangeSet

	for _, c := range cs.Changes {
		var change provider.Change

		switch c.Action {
		case types.ChangeActionCreate:
			change = provider.Created
		case types.ChangeActionDelete:
			change = provider.Deleted
		case types.ChangeActionUpsert:
			change = provider.Updated
		}

		switch c.ResourceRecordSet.Type {
		case types.RRTypePtr:
			result.PTR = change
		case types.RRTypeSrv:
			result.SRV = change
		case types.RRTypeTxt:
			result.TXT = change
		}

		for _, rec := range c.ResourceRecordSet.ResourceRecords {
			a.Logger.Info(
				string(c.Action)+" record",
				"type", c.ResourceRecordSet.Type,
				"name", c.ResourceRecordSet.Name,
				"value", *rec.Value,
				"ttl", *c.ResourceRecordSet.TTL,
			)
		}
	}

	return result, nil
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
		dnssd.ServiceInstanceName(inst.Name, inst.ServiceType, inst.Domain) + ".",
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
