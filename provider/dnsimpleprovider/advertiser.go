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
) (provider.AdvertiseResult, error) {
	cs := &changeSet{}

	existingPTR, err := a.syncPTR(ctx, inst, cs)
	if err != nil {
		return provider.AdvertiseError, err
	}

	existingSRV, err := a.syncSRV(ctx, inst, cs)
	if err != nil {
		return provider.AdvertiseError, err
	}

	existingTXT, err := a.syncTXT(ctx, inst, cs)
	if err != nil {
		return provider.AdvertiseError, err
	}

	hasChanges, err := a.applyChangeSet(ctx, cs)
	if err != nil {
		return provider.AdvertiseError, err
	}

	if existingPTR && existingSRV && existingTXT {
		if hasChanges {
			return provider.UpdatedExistingInstance, nil
		}

		return provider.InstanceAlreadyAdvertised, nil
	}

	return provider.AdvertisedNewInstance, nil
}

func (a *advertiser) Unadvertise(
	ctx context.Context,
	inst dnssd.ServiceInstance,
) (provider.UnadvertiseResult, error) {
	cs := &changeSet{}

	if err := a.deletePTR(ctx, inst, cs); err != nil {
		return provider.UnadvertiseError, err
	}

	if err := a.deleteSRV(ctx, inst, cs); err != nil {
		return provider.UnadvertiseError, err
	}

	if err := a.deleteTXT(ctx, inst, cs); err != nil {
		return provider.UnadvertiseError, err
	}

	hasChanges, err := a.applyChangeSet(ctx, cs)
	if err != nil {
		return provider.UnadvertiseError, err
	}

	if hasChanges {
		return provider.UnadvertisedExistingInstance, nil
	}

	return provider.InstanceNotAdvertised, nil
}

func (a *advertiser) applyChangeSet(
	ctx context.Context,
	cs *changeSet,
) (bool, error) {
	accountID := strconv.FormatInt(a.Zone.AccountID, 10)
	ok := false

	for _, rec := range cs.deletes {
		ok = true
		if _, err := a.API.DeleteRecord(ctx, accountID, a.Zone.Name, rec.ID); err != nil {
			return false, fmt.Errorf("unable to delete %s record: %w", rec.Type, err)
		}
	}

	for _, up := range cs.updates {
		ok = true
		if _, err := a.API.UpdateRecord(ctx, accountID, a.Zone.Name, up.Before.ID, up.After); err != nil {
			return false, fmt.Errorf("unable to update %s record: %w", up.Before.Type, err)
		}
	}

	for _, attr := range cs.creates {
		ok = true
		if _, err := a.API.CreateRecord(ctx, accountID, a.Zone.Name, attr); err != nil {
			return false, fmt.Errorf("unable to create %s record: %w", attr.Type, err)
		}
	}

	return ok, nil
}
