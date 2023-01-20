package dnsimpleprovider

import (
	"context"
	"fmt"
	"strings"

	"github.com/dnsimple/dnsimple-go/dnsimple"
	"github.com/dogmatiq/dissolve/dnssd"
	"golang.org/x/exp/slices"
)

type advertiser struct {
	API          *dnsimple.ZonesService
	AdvertiserID string
	AccountID    string
	ZoneID       string
}

func (a *advertiser) ID() string {
	return a.AdvertiserID
}

func (a *advertiser) Advertise(
	ctx context.Context,
	inst dnssd.ServiceInstance,
) error {
	qualifiedInstanceName := dnssd.ServiceInstanceName(inst.Instance, inst.ServiceType, inst.Domain)

	instanceName := dnsimple.String(
		dnssd.EscapeInstance(inst.Instance) + "." + inst.ServiceType,
	)

	serviceName := dnsimple.String(
		inst.ServiceType,
	)

	create := []dnsimple.ZoneRecordAttributes{
		{
			ZoneID:  a.ZoneID,
			Type:    "PTR",
			Name:    serviceName,
			Content: qualifiedInstanceName,
			TTL:     int(inst.TTL.Seconds()),
		},
		{
			ZoneID: a.ZoneID,
			Type:   "SRV",
			Name:   instanceName,
			Content: fmt.Sprintf(
				"%d %d %s",
				inst.Weight,
				inst.TargetPort,
				inst.TargetHost,
			),
			TTL:      int(inst.TTL.Seconds()),
			Priority: int(inst.Priority),
		},
	}

	for _, r := range dnssd.NewTXTRecords(inst) {
		create = append(
			create,
			dnsimple.ZoneRecordAttributes{
				ZoneID:  a.ZoneID,
				Type:    "TXT",
				Name:    instanceName,
				Content: strings.TrimPrefix(r.String(), r.Hdr.String()),
				TTL:     int(inst.TTL.Seconds()),
			},
		)
	}

	if err := forEach(
		ctx,
		func(opts dnsimple.ListOptions) ([]dnsimple.ZoneRecord, error) {
			res, err := a.API.ListRecords(
				ctx,
				a.AccountID,
				a.ZoneID,
				&dnsimple.ZoneRecordListOptions{
					ListOptions: opts,
					Name:        serviceName,
					Type:        dnsimple.String("PTR"),
				},
			)
			if err != nil {
				return nil, fmt.Errorf("unable to list zone records: %w", err)
			}

			return res.Data, nil
		},
		func(r dnsimple.ZoneRecord) (bool, error) {
			for i, attr := range create {
				if r.Content != qualifiedInstanceName {
					continue
				}

				if recordHasMatchingAttributes(r, attr) {
					create = slices.Delete(create, i, i+1)
					return false, nil
				}

				return false, a.deleteRecord(ctx, r)
			}

			return true, nil
		},
	); err != nil {
		return err
	}

	if err := forEach(
		ctx,
		func(opts dnsimple.ListOptions) ([]dnsimple.ZoneRecord, error) {
			res, err := a.API.ListRecords(
				ctx,
				a.AccountID,
				a.ZoneID,
				&dnsimple.ZoneRecordListOptions{
					ListOptions: opts,
					Name:        instanceName,
				},
			)
			if err != nil {
				return nil, fmt.Errorf("unable to list zone records: %w", err)
			}

			return res.Data, nil
		},
		func(r dnsimple.ZoneRecord) (bool, error) {
			for i, attr := range create {
				if recordHasMatchingAttributes(r, attr) {
					create = slices.Delete(create, i, i+1)
					return true, nil
				}
			}

			return true, a.deleteRecord(ctx, r)
		},
	); err != nil {
		return err
	}

	for _, r := range create {
		if _, err := a.API.CreateRecord(ctx, a.AccountID, a.ZoneID, r); err != nil {
			return fmt.Errorf("unable to create zone record: %w", err)
		}
	}

	return nil
}

func (a *advertiser) Unadvertise(
	ctx context.Context,
	inst dnssd.ServiceInstance,
) error {
	qualifiedInstanceName := dnssd.ServiceInstanceName(inst.Instance, inst.ServiceType, inst.Domain)

	instanceName := dnsimple.String(
		dnssd.EscapeInstance(inst.Instance) + "." + inst.ServiceType,
	)

	serviceName := dnsimple.String(
		inst.ServiceType,
	)

	if err := forEach(
		ctx,
		func(opts dnsimple.ListOptions) ([]dnsimple.ZoneRecord, error) {
			res, err := a.API.ListRecords(
				ctx,
				a.AccountID,
				a.ZoneID,
				&dnsimple.ZoneRecordListOptions{
					ListOptions: opts,
					Name:        serviceName,
					Type:        dnsimple.String("PTR"),
				},
			)
			if err != nil {
				return nil, fmt.Errorf("unable to list zone records: %w", err)
			}

			return res.Data, nil
		},
		func(r dnsimple.ZoneRecord) (bool, error) {
			if r.Content == qualifiedInstanceName {
				return false, a.deleteRecord(ctx, r)
			}

			return true, nil
		},
	); err != nil {
		return err
	}

	if err := forEach(
		ctx,
		func(opts dnsimple.ListOptions) ([]dnsimple.ZoneRecord, error) {
			res, err := a.API.ListRecords(
				ctx,
				a.AccountID,
				a.ZoneID,
				&dnsimple.ZoneRecordListOptions{
					ListOptions: opts,
					Name:        instanceName,
				},
			)
			if err != nil {
				return nil, fmt.Errorf("unable to list zone records: %w", err)
			}

			return res.Data, nil
		},
		func(r dnsimple.ZoneRecord) (bool, error) {
			return true, a.deleteRecord(ctx, r)
		},
	); err != nil {
		return err
	}

	return nil
}

func (a *advertiser) deleteRecord(ctx context.Context, r dnsimple.ZoneRecord) error {
	if _, err := a.API.DeleteRecord(ctx, a.AccountID, a.ZoneID, r.ID); err != nil {
		return fmt.Errorf("unable to delete zone record: %w", err)
	}

	return nil
}
