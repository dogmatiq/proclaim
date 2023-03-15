package dnsimpleprovider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/dnsimple/dnsimple-go/dnsimple"
	"github.com/dogmatiq/dissolve/dnssd"
	"github.com/dogmatiq/proclaim/provider/dnsimpleprovider/internal/dnsimplex"
)

func (a *advertiser) findSRV(
	ctx context.Context,
	inst dnssd.ServiceInstance,
) (dnsimple.ZoneRecord, bool, error) {
	return dnsimplex.One(
		ctx,
		func(opts dnsimple.ListOptions) (*dnsimple.Pagination, []dnsimple.ZoneRecord, error) {
			res, err := a.API.ListRecords(
				ctx,
				strconv.FormatInt(a.Zone.AccountID, 10),
				a.Zone.Name,
				&dnsimple.ZoneRecordListOptions{
					ListOptions: opts,
					Name: dnsimple.String(
						dnssd.EscapeInstance(inst.Instance) + "." + inst.ServiceType,
					),
					Type: dnsimple.String("SRV"),
				},
			)
			if err != nil {
				return nil, nil, fmt.Errorf("unable to list SRV records: %w", err)
			}

			return res.Pagination, res.Data, nil
		},
	)
}

func (a *advertiser) syncSRV(
	ctx context.Context,
	inst dnssd.ServiceInstance,
	cs *changeSet,
) error {
	current, ok, err := a.findSRV(ctx, inst)
	if err != nil {
		return err
	}

	desired := dnsimple.ZoneRecordAttributes{
		ZoneID: a.Zone.Name,
		Type:   "SRV",
		Name: dnsimple.String(
			dnssd.EscapeInstance(inst.Instance) + "." + inst.ServiceType,
		),
		Content: fmt.Sprintf(
			"%d %d %s",
			inst.Weight,
			inst.TargetPort,
			inst.TargetHost,
		),
		TTL:      int(inst.TTL.Seconds()),
		Priority: int(inst.Priority),
	}

	if ok {
		cs.Update(current, desired)
	} else {
		cs.Create(desired)
	}

	return nil
}

func (a *advertiser) deleteSRV(
	ctx context.Context,
	inst dnssd.ServiceInstance,
	cs *changeSet,
) error {
	current, ok, err := a.findSRV(ctx, inst)
	if !ok || err != nil {
		return err
	}

	cs.Delete(current)

	return nil
}
