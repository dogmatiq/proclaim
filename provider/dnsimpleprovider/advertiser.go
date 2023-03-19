package dnsimpleprovider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/dnsimple/dnsimple-go/dnsimple"
	"github.com/dogmatiq/dissolve/dnssd"
	"github.com/dogmatiq/proclaim/provider"
)

type advertiser struct {
	API  *dnsimple.ZonesService
	Zone *dnsimple.Zone
}

func (a *advertiser) ID() string {
	return marshalAdvertiserID(a.Zone)
}

func (a *advertiser) Advertise(
	ctx context.Context,
	inst dnssd.ServiceInstance,
) (provider.ChangeSet, error) {
	cs := &changeSet{}

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
	cs := &changeSet{}

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
	cs *changeSet,
) (provider.ChangeSet, error) {
	var result provider.ChangeSet
	accountID := strconv.FormatInt(a.Zone.AccountID, 10)

	for _, rec := range cs.deletes {
		if _, err := a.API.DeleteRecord(ctx, accountID, a.Zone.Name, rec.ID); err != nil {
			return provider.ChangeSet{}, fmt.Errorf("unable to delete %s record: %w", rec.Type, err)
		}

		switch rec.Type {
		case "PTR":
			result.PTR |= provider.Deleted
		case "SRV":
			result.SRV |= provider.Deleted
		case "TXT":
			result.TXT |= provider.Deleted
		}
	}

	for _, up := range cs.updates {
		if _, err := a.API.UpdateRecord(ctx, accountID, a.Zone.Name, up.Before.ID, up.After); err != nil {
			return provider.ChangeSet{}, fmt.Errorf("unable to update %s record: %w", up.Before.Type, err)
		}

		switch up.Before.Type {
		case "PTR":
			result.PTR |= provider.Updated
		case "SRV":
			result.SRV |= provider.Updated
		case "TXT":
			result.TXT |= provider.Updated
		}
	}

	for _, attr := range cs.creates {
		if _, err := a.API.CreateRecord(ctx, accountID, a.Zone.Name, attr); err != nil {
			return provider.ChangeSet{}, fmt.Errorf("unable to create %s record: %w", attr.Type, err)
		}

		switch attr.Type {
		case "PTR":
			result.PTR |= provider.Created
		case "SRV":
			result.SRV |= provider.Created
		case "TXT":
			result.TXT |= provider.Created
		}
	}

	return result, nil
}
