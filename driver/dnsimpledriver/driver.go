package dnsimpledriver

import (
	"context"
	"fmt"
	"strconv"

	"github.com/dnsimple/dnsimple-go/dnsimple"
	"github.com/dogmatiq/proclaim"
	"github.com/go-logr/logr"
)

// Driver is an implementation of proclaim.Driver that advertises DNS-SD
// services on domains hosted by dnsimple.com.
type Driver struct {
	API *dnsimple.Client
}

// Name returns a human-readable name for the driver.
func (d *Driver) Name() string {
	return "DNSimple"
}

// AdvertiserForDomain returns the Advertiser used to advertise services on the
// given domain.
//
// ok is false if this driver does not manage the given domain.
func (d *Driver) AdvertiserForDomain(
	ctx context.Context,
	logger logr.Logger,
	domain string,
) (adv proclaim.Advertiser, ok bool, _ error) {
	err := forEach(
		ctx,
		func(opts dnsimple.ListOptions) ([]dnsimple.Account, error) {
			res, err := d.API.Accounts.ListAccounts(ctx, &opts)
			if err != nil {
				return nil, fmt.Errorf("unable to list accounts: %w", err)
			}

			return res.Data, nil
		},
		func(a dnsimple.Account) (bool, error) {
			accountID := strconv.FormatInt(a.ID, 10)
			res, err := d.API.Zones.GetZone(ctx, accountID, domain)
			if err != nil {
				return false, fmt.Errorf("unable to get %q zone: %w", domain, err)
			}

			if res.Data == nil {
				return true, nil
			}

			ok = true
			adv = &advertiser{
				API:       d.API.Zones,
				AccountID: accountID,
				ZoneID:    strconv.FormatInt(res.Data.ID, 10),
			}

			return false, nil
		},
	)

	return adv, ok, err
}
