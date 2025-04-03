package route53provider

import route53advertiser "github.com/dogmatiq/dissolve/dnssd/advertiser/route53"

type advertiser struct {
	*route53advertiser.Advertiser
	ZoneID string
}

func (a *advertiser) ID() map[string]any {
	return marshalAdvertiserID(a.ZoneID)
}
