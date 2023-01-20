package dnsimpleprovider

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/dnsimple/dnsimple-go/dnsimple"
	"github.com/dogmatiq/proclaim/provider"
)

// Provider is an implementation of provider.Provider that advertises DNS-SD
// services on domains hosted by dnsimple.com.
type Provider struct {
	API *dnsimple.Client

	once sync.Once
	id   string
}

// ID returns a short unique identifier for the provider.
func (p *Provider) ID() string {
	p.once.Do(func() {
		p.id = "dnsimple"

		if p.API.BaseURL != "" {
			u, err := url.Parse(p.API.BaseURL)
			if err != nil {
				panic(err)
			}

			if u.Host != "api.dnsimple.com" {
				p.id += "/" + u.Host
			}
		}
	})

	return p.id
}

// AdvertiserByID returns the Advertiser with the given ID.
func (p *Provider) AdvertiserByID(ctx context.Context, id string) (provider.Advertiser, error) {
	parts := strings.SplitN(id, "/", 2)
	if len(parts) != 2 {
		return nil, errors.New("invalid advertiser ID")
	}

	accountID, domain := parts[0], parts[1]
	res, err := p.API.Zones.GetZone(ctx, accountID, domain)
	if err != nil {
		return nil, fmt.Errorf(
			"unable to get %q zone on account %q: %w",
			domain,
			accountID,
			err,
		)
	}

	if res.Data == nil {
		return nil, fmt.Errorf(
			"unable to get %q zone on account %q: no such zone",
			domain,
			accountID,
		)
	}

	return p.newAdvertiser(accountID, domain, res), nil
}

// AdvertiserByDomain returns the Advertiser used to advertise services on the
// given domain.
//
// ok is false if this provider does not manage the given domain.
func (p *Provider) AdvertiserByDomain(ctx context.Context, domain string) (adv provider.Advertiser, ok bool, _ error) {
	err := forEach(
		ctx,
		func(opts dnsimple.ListOptions) ([]dnsimple.Account, error) {
			res, err := p.API.Accounts.ListAccounts(ctx, &opts)
			if err != nil {
				return nil, fmt.Errorf("unable to list accounts: %w", err)
			}

			return res.Data, nil
		},
		func(a dnsimple.Account) (bool, error) {
			accountID := strconv.FormatInt(a.ID, 10)
			res, err := p.API.Zones.GetZone(ctx, accountID, domain)
			if err != nil {
				return false, fmt.Errorf(
					"unable to get %q zone on account %q: %w",
					domain,
					accountID,
					err,
				)
			}

			if res.Data == nil {
				return true, nil
			}

			ok = true
			adv = p.newAdvertiser(accountID, domain, res)

			return false, nil
		},
	)

	return adv, ok, err
}

func (p *Provider) newAdvertiser(
	accountID, domain string,
	zone *dnsimple.ZoneResponse,
) provider.Advertiser {
	return &advertiser{
		API:          p.API.Zones,
		AdvertiserID: fmt.Sprintf("%s/%s", accountID, domain),
		AccountID:    accountID,
		ZoneID:       strconv.FormatInt(zone.Data.ID, 10),
	}
}
