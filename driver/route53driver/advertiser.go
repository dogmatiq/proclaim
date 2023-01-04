package route53driver

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/route53/route53iface"
	"github.com/dogmatiq/dissolve/dnssd"
	"github.com/go-logr/logr"
)

type advertiser struct {
	API    route53iface.Route53API
	ZoneID *string
}

func (a *advertiser) Advertise(
	ctx context.Context,
	logger logr.Logger,
	inst dnssd.ServiceInstance,
) error {
	instanceName := aws.String(
		dnssd.ServiceInstanceName(inst.Instance, inst.ServiceType, inst.Domain) + ".",
	)

	serviceName := aws.String(
		dnssd.InstanceEnumerationDomain(inst.ServiceType, inst.Domain) + ".",
	)

	ttl := aws.Int64(int64(inst.TTL.Seconds()))

	ptr, err := a.getResourceRecordSet(ctx, serviceName, "PTR")
	if err != nil {
		return err
	}

	changes, err := addInstanceToPTRChanges(ptr, instanceName, serviceName)
	if err != nil {
		return err
	}

	changes = append(
		changes,
		&route53.Change{
			Action: aws.String(route53.ChangeActionUpsert),
			ResourceRecordSet: &route53.ResourceRecordSet{
				Type: aws.String("SRV"),
				Name: instanceName,
				TTL:  ttl,
				ResourceRecords: convertRecords(
					dnssd.NewSRVRecord(inst),
				),
			},
		},
		&route53.Change{
			Action: aws.String(route53.ChangeActionUpsert),
			ResourceRecordSet: &route53.ResourceRecordSet{
				Type: aws.String("TXT"),
				Name: instanceName,
				TTL:  ttl,
				ResourceRecords: convertRecords(
					dnssd.NewTXTRecords(inst)...,
				),
			},
		},
	)

	_, err = a.API.ChangeResourceRecordSetsWithContext(
		ctx,
		&route53.ChangeResourceRecordSetsInput{
			HostedZoneId: a.ZoneID,
			ChangeBatch: &route53.ChangeBatch{
				Comment: aws.String(fmt.Sprintf(
					"dogmatiq/proclaim: advertising %s instance: %s ",
					inst.ServiceType,
					inst.Instance,
				)),
				Changes: changes,
			},
		},
	)
	return err
}

func (a *advertiser) Unadvertise(
	ctx context.Context,
	logger logr.Logger,
	inst dnssd.ServiceInstance,
) error {
	instanceName := aws.String(
		dnssd.ServiceInstanceName(inst.Instance, inst.ServiceType, inst.Domain) + ".",
	)

	serviceName := aws.String(
		dnssd.InstanceEnumerationDomain(inst.ServiceType, inst.Domain) + ".",
	)

	ptr, err := a.getResourceRecordSet(ctx, serviceName, "PTR")
	if err != nil {
		return err
	}

	changes, err := removeInstanceFromPTRChanges(ptr, instanceName, serviceName)
	if err != nil {
		return err
	}

	if err := a.API.ListResourceRecordSetsPagesWithContext(
		ctx,
		&route53.ListResourceRecordSetsInput{
			HostedZoneId:    a.ZoneID,
			StartRecordName: instanceName,
			StartRecordType: aws.String("SRV"),
		},
		func(res *route53.ListResourceRecordSetsOutput, lastPage bool) bool {
			for _, set := range res.ResourceRecordSets {
				if *set.Name != *instanceName {
					return false
				}

				if *set.Type > "TXT" {
					return false
				}

				if *set.Type != "SRV" && *set.Type != "TXT" {
					continue
				}

				changes = append(
					changes,
					&route53.Change{
						Action:            aws.String(route53.ChangeActionDelete),
						ResourceRecordSet: set,
					},
				)
			}

			return true
		},
	); err != nil {
		return fmt.Errorf("unable to list resource record sets: %w", err)
	}

	if len(changes) == 0 {
		logger.Info("no DNS changes to be made")
		return nil
	}

	_, err = a.API.ChangeResourceRecordSetsWithContext(
		ctx,
		&route53.ChangeResourceRecordSetsInput{
			HostedZoneId: a.ZoneID,
			ChangeBatch: &route53.ChangeBatch{
				Comment: aws.String(fmt.Sprintf(
					"dogmatiq/proclaim: unadvertising %s instance: %s ",
					inst.ServiceType,
					inst.Instance,
				)),
				Changes: changes,
			},
		},
	)
	return err
}

func (a *advertiser) getResourceRecordSet(
	ctx context.Context,
	name *string,
	recordType string,
) (*route53.ResourceRecordSet, error) {
	res, err := a.API.ListResourceRecordSetsWithContext(
		ctx,
		&route53.ListResourceRecordSetsInput{
			HostedZoneId:    a.ZoneID,
			StartRecordName: name,
			StartRecordType: aws.String(recordType),
			MaxItems:        aws.String("1"),
		},
	)
	if err != nil {
		return nil, err
	}

	if len(res.ResourceRecordSets) == 0 {
		return nil, nil
	}

	set := res.ResourceRecordSets[0]

	if !strings.EqualFold(*set.Name, *name) {
		return nil, nil
	}

	if !strings.EqualFold(*set.Type, recordType) {
		return nil, nil
	}

	return set, nil
}
