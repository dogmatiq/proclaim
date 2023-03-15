package dnsimpleprovider

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/dnsimple/dnsimple-go/dnsimple"
	"github.com/dogmatiq/proclaim/provider"
	"github.com/dogmatiq/proclaim/provider/dnsimpleprovider/internal/dnsimplex"
)

// Provider is an implementation of provider.Provider that advertises DNS-SD
// services on domains hosted by dnsimple.com.
type Provider struct {
	API *dnsimple.Client
}

// ID returns a short unique identifier for the provider.
func (p *Provider) ID() string {
	return base64.RawURLEncoding.EncodeToString(
		[]byte(p.API.BaseURL),
	)
}

// Describe returns a human-readable description of the provider.
func (p *Provider) Describe() string {
	u, err := url.Parse(p.API.BaseURL)
	if err != nil {
		panic(err)
	}

	if u.Host == "api.dnsimple.com" {
		return "dnsimple"
	}

	environment := strings.TrimPrefix(u.Host, "api.")
	environment = strings.TrimSuffix(environment, ".dnsimple.com")

	return fmt.Sprintf("dnsimple (%s)", environment)
}

// AdvertiserByID returns the Advertiser with the given ID.
func (p *Provider) AdvertiserByID(
	ctx context.Context,
	id string,
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
			res, err := p.API.Accounts.ListAccounts(ctx, &opts)
			if err != nil {
				return nil, nil, fmt.Errorf("unable to list accounts: %w", err)
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
	res, err := p.API.Zones.GetZone(
		ctx,
		strconv.FormatInt(accountID, 10),
		domain,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"unable to get %q zone on account %d: %w",
			domain,
			accountID,
			err,
		)
	}

	return &advertiser{
		API:  p.API.Zones,
		Zone: res.Data,
	}, nil
}

// marshalAdvertiserID returns the ID of the advertiser for the given zone.
func marshalAdvertiserID(z *dnsimple.Zone) string {
	var data []byte
	data = binary.BigEndian.AppendUint64(data, uint64(z.AccountID))
	data = append(data, z.Name...)
	return base64.RawURLEncoding.EncodeToString(data)
}

// unmarshalAdvertiserID parses an advertiser ID into its constituent parts.
func unmarshalAdvertiserID(id string) (accountID int64, domain string, err error) {
	data, err := base64.RawURLEncoding.DecodeString(id)
	if err != nil {
		return 0, "", fmt.Errorf("invalid advertiser ID: %w", err)
	}

	if len(data) < 8 {
		return 0, "", errors.New("invalid advertiser ID: too short")
	}

	accountID = int64(binary.BigEndian.Uint64(data))
	if accountID <= 0 {
		return 0, "", errors.New("invalid advertiser ID: account ID component must be positive")
	}

	domain = string(data[8:])
	if domain == "" {
		return 0, "", errors.New("invalid advertiser ID: domain component must not be empty")
	}

	return accountID, domain, nil
}
