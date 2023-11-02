package dnsimpleprovider

import (
	"context"
	"strconv"

	"github.com/dnsimple/dnsimple-go/dnsimple"
	"github.com/dogmatiq/dissolve/dnssd"
	"github.com/dogmatiq/proclaim/provider"
	"github.com/dogmatiq/proclaim/provider/dnsimpleprovider/internal/dnsimplex"
	"github.com/go-logr/logr"
)

type advertiser struct {
	Client *dnsimple.ZonesService
	Zone   *dnsimple.Zone
	Logger logr.Logger
}

func (a *advertiser) ID() map[string]any {
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
		if _, err := a.Client.DeleteRecord(ctx, accountID, a.Zone.Name, rec.ID); err != nil {
			return provider.ChangeSet{}, dnsimplex.Errorf("unable to delete %s record: %w", rec.Type, err)
		}

		switch rec.Type {
		case "PTR":
			result.PTR |= provider.Deleted
		case "SRV":
			result.SRV |= provider.Deleted
		case "TXT":
			result.TXT |= provider.Deleted
		}

		a.Logger.Info(
			"DELETE record",
			"type", rec.Type,
			"name", rec.Name,
			"content", rec.Content,
			"priority", rec.Priority,
			"ttl", rec.TTL,
		)
	}

	for _, up := range cs.updates {
		if _, err := a.Client.UpdateRecord(ctx, accountID, a.Zone.Name, up.Before.ID, up.After); err != nil {
			return provider.ChangeSet{}, dnsimplex.Errorf("unable to update %s record: %w", up.Before.Type, err)
		}

		switch up.Before.Type {
		case "PTR":
			result.PTR |= provider.Updated
		case "SRV":
			result.SRV |= provider.Updated
		case "TXT":
			result.TXT |= provider.Updated
		}

		a.Logger.Info(
			"UPDATE record",
			"type", up.Before.Type,
			"name", up.Before.Name,
			"content_before", up.Before.Content,
			"priority_before", up.Before.Priority,
			"ttl_before", up.Before.TTL,
			"content_after", up.After.Content,
			"priority_after", up.After.Priority,
			"ttl_after", up.After.TTL,
		)
	}

	for _, attr := range cs.creates {
		if _, err := a.Client.CreateRecord(ctx, accountID, a.Zone.Name, attr); err != nil {
			return provider.ChangeSet{}, dnsimplex.Errorf("unable to create %s record: %w", attr.Type, err)
		}

		switch attr.Type {
		case "PTR":
			result.PTR |= provider.Created
		case "SRV":
			result.SRV |= provider.Created
		case "TXT":
			result.TXT |= provider.Created
		}

		a.Logger.Info(
			"CREATE record",
			"type", attr.Type,
			"name", attr.Name,
			"content", attr.Content,
			"priority", attr.Priority,
			"ttl", attr.TTL,
		)
	}

	return result, nil
}
