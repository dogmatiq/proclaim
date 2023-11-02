package dnsimpleprovider

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/dnsimple/dnsimple-go/dnsimple"
	"github.com/dogmatiq/proclaim/provider"
	"github.com/dogmatiq/proclaim/provider/dnsimpleprovider/internal/dnsimplex"
	"github.com/go-logr/logr"
)

// Provider is an implementation of provider.Provider that advertises DNS-SD
// services on domains hosted by dnsimple.com.
type Provider struct {
	Client *dnsimple.Client
	Logger logr.Logger
}

// ID returns a short unique identifier for the provider.
func (p *Provider) ID() string {
	if env := p.environment(); env != "production" {
		return fmt.Sprintf("dnsimple.%s", env)
	}
	return "dnsimple"
}

func (p *Provider) environment() string {
	u, err := url.Parse(p.Client.BaseURL)
	if err != nil {
		panic(err)
	}

	if u.Host == "api.dnsimple.com" {
		return "production"
	}

	environment := strings.TrimPrefix(u.Host, "api.")
	environment = strings.TrimSuffix(environment, ".dnsimple.com")
	return environment
}

// Describe returns a human-readable description of the provider.
func (p *Provider) Describe() string {
	if env := p.environment(); env != "production" {
		return fmt.Sprintf("DNSimple (%s)", env)
	}
	return "DNSimple"
}

// AdvertiserByID returns the Advertiser with the given ID.
func (p *Provider) AdvertiserByID(
	ctx context.Context,
	id map[string]any,
) (provider.Advertiser, error) {
	accountID, domain, err := unmarshalAdvertiserID(id)
	if err != nil {
		return nil, err
	}

	return p.advertiserByDomain(ctx, accountID, domain)
}

// AdvertiserByDomain returns the Advertiser used to advertise services on the
// given domain.
//
// ok is false if this provider does not manage the given domain.
func (p *Provider) AdvertiserByDomain(
	ctx context.Context,
	domain string,
) (provider.Advertiser, bool, error) {
	return dnsimplex.Find(
		ctx,
		func(opts dnsimple.ListOptions) (*dnsimple.Pagination, []dnsimple.Account, error) {
			res, err := p.Client.Accounts.ListAccounts(ctx, &opts)
			if err != nil {
				return nil, nil, dnsimplex.Errorf("unable to list accounts: %w", err)
			}
			return res.Pagination, res.Data, err
		},
		func(acc dnsimple.Account) (provider.Advertiser, bool, error) {
			a, err := p.advertiserByDomain(ctx, acc.ID, domain)
			return a, err == nil, dnsimplex.IgnoreNotFound(err)
		},
	)
}

// advertiserByDomain returns the Advertiser used to advertise services on the
// given domain under the given account.
func (p *Provider) advertiserByDomain(
	ctx context.Context,
	accountID int64,
	domain string,
) (provider.Advertiser, error) {
	res, err := p.Client.Zones.GetZone(
		ctx,
		strconv.FormatInt(accountID, 10),
		domain,
	)
	if err != nil {
		return nil, dnsimplex.Errorf(
			"unable to get %q zone on account %d: %w",
			domain,
			accountID,
			err,
		)
	}

	return &advertiser{
		p.Client.Zones,
		res.Data,
		p.Logger,
	}, nil
}

// marshalAdvertiserID returns the ID of the advertiser for the given zone.
func marshalAdvertiserID(z *dnsimple.Zone) map[string]any {
	return map[string]any{
		"accountID": z.AccountID,
		"zoneName":  z.Name,
	}
}

// unmarshalAdvertiserID parses an advertiser ID into its constituent parts.
func unmarshalAdvertiserID(id map[string]any) (accountID int64, domain string, err error) {
	accountIDAny, ok := id["accountID"]
	if !ok {
		return 0, "", errors.New("invalid advertiser ID: missing accountID key")
	}

	accountID, ok = accountIDAny.(int64)
	if !ok || accountID <= 0 {
		return 0, "", errors.New("invalid advertiser ID: accountID must be a positive integer")
	}

	zoneNameAny, ok := id["zoneName"]
	if !ok {
		return 0, "", errors.New("invalid advertiser ID: missing zoneName key")
	}

	domain, ok = zoneNameAny.(string)
	if !ok || domain == "" {
		return 0, "", errors.New("invalid advertiser ID: zoneName must be a non-empty string")
	}

	return accountID, domain, nil
}
