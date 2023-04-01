package dnsimpleprovider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/dnsimple/dnsimple-go/dnsimple"
	"github.com/dogmatiq/dissolve/dnssd"
	"github.com/dogmatiq/proclaim/provider/dnsimpleprovider/internal/dnsimplex"
	"golang.org/x/exp/slices"
)

func (a *advertiser) findTXT(
	ctx context.Context,
	name dnssd.ServiceInstanceName,
) ([]dnsimple.ZoneRecord, error) {
	return dnsimplex.All(
		ctx,
		func(opts dnsimple.ListOptions) (*dnsimple.Pagination, []dnsimple.ZoneRecord, error) {
			res, err := a.Client.ListRecords(
				ctx,
				strconv.FormatInt(a.Zone.AccountID, 10),
				a.Zone.Name,
				&dnsimple.ZoneRecordListOptions{
					ListOptions: opts,
					Name:        dnsimple.String(name.Relative()),
					Type:        dnsimple.String("TXT"),
				},
			)
			if err != nil {
				return nil, nil, fmt.Errorf("unable to list TXT records: %w", err)
			}

			return res.Pagination, res.Data, nil
		},
	)
}

func (a *advertiser) syncTXT(
	ctx context.Context,
	inst dnssd.ServiceInstance,
	cs *changeSet,
) error {
	current, err := a.findTXT(ctx, inst.ServiceInstanceName)
	if err != nil {
		return err
	}

	var desired []dnsimple.ZoneRecordAttributes

	for _, r := range dnssd.NewTXTRecords(inst) {
		desired = append(
			desired,
			dnsimple.ZoneRecordAttributes{
				ZoneID:  a.Zone.Name,
				Type:    "TXT",
				Name:    dnsimple.String(inst.Relative()),
				Content: strings.TrimPrefix(r.String(), r.Hdr.String()),
				TTL:     int(inst.TTL.Seconds()),
			},
		)
	}

next:
	for _, c := range current {
		for i, d := range desired {
			if c.Content == d.Content {
				// We consider a TXT record with the same content to be the same
				// record.
				desired = slices.Delete(desired, i, i+1)
				cs.Update(c, d)
				continue next
			}
		}

		cs.Delete(c)
	}

	for _, attr := range desired {
		cs.Create(attr)
	}

	return nil
}

func (a *advertiser) deleteTXT(
	ctx context.Context,
	name dnssd.ServiceInstanceName,
	cs *changeSet,
) error {
	current, err := a.findTXT(ctx, name)
	if err != nil {
		return err
	}

	for _, c := range current {
		cs.Delete(c)
	}

	return nil
}
