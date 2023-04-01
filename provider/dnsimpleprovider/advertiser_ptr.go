package dnsimpleprovider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/dnsimple/dnsimple-go/dnsimple"
	"github.com/dogmatiq/dissolve/dnssd"
	"github.com/dogmatiq/proclaim/provider/dnsimpleprovider/internal/dnsimplex"
)

func (a *advertiser) findPTR(
	ctx context.Context,
	name dnssd.ServiceInstanceName,
) (dnsimple.ZoneRecord, bool, error) {
	return dnsimplex.First(
		ctx,
		func(opts dnsimple.ListOptions) (*dnsimple.Pagination, []dnsimple.ZoneRecord, error) {
			res, err := a.Client.ListRecords(
				ctx,
				strconv.FormatInt(a.Zone.AccountID, 10),
				a.Zone.Name,
				&dnsimple.ZoneRecordListOptions{
					ListOptions: opts,
					Name:        dnsimple.String(name.ServiceType),
					Type:        dnsimple.String("PTR"),
				},
			)
			if err != nil {
				return nil, nil, fmt.Errorf("unable to list PTR records: %w", err)
			}

			return res.Pagination, res.Data, nil
		},
		func(candidate dnsimple.ZoneRecord) bool {
			return candidate.Content == dnssd.AbsoluteServiceInstanceName(name.Name, name.ServiceType, name.Domain)
		},
	)
}

func (a *advertiser) syncPTR(
	ctx context.Context,
	inst dnssd.ServiceInstance,
	cs *changeSet,
) error {
	current, ok, err := a.findPTR(ctx, inst.ServiceInstanceName)
	if err != nil {
		return err
	}

	desired := dnsimple.ZoneRecordAttributes{
		ZoneID:  a.Zone.Name,
		Type:    "PTR",
		Name:    dnsimple.String(inst.ServiceType),
		Content: dnssd.AbsoluteServiceInstanceName(inst.Name, inst.ServiceType, inst.Domain),
		TTL:     int(inst.TTL.Seconds()),
	}

	if ok {
		cs.Update(current, desired)
	} else {
		cs.Create(desired)
	}

	return nil
}

func (a *advertiser) deletePTR(
	ctx context.Context,
	name dnssd.ServiceInstanceName,
	cs *changeSet,
) error {
	current, ok, err := a.findPTR(ctx, name)
	if !ok || err != nil {
		return err
	}

	cs.Delete(current)

	return nil
}
