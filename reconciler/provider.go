package reconciler

import (
	"context"
	"fmt"

	"github.com/dogmatiq/proclaim/crd"
	"github.com/dogmatiq/proclaim/provider"
	"golang.org/x/exp/slices"
)

// getOrAssociateAdvertiser returns the advertiser used to
// advertise/un-advertise the given DNS-SD service instance.
func (r *Reconciler) getOrAssociateAdvertiser(
	ctx context.Context,
	res *crd.DNSSDServiceInstance,
) (provider.Advertiser, bool, error) {
	if res.Status.ProviderID == "" {
		return r.associateAdvertiser(ctx, res)
	}

	return r.getAdvertiser(ctx, res)
}

// associateAdvertiser finds the appropriate advertiser for the given DNS-SD
// service instance from all available providers and associates it with the
// resource.
func (r *Reconciler) associateAdvertiser(
	ctx context.Context,
	res *crd.DNSSDServiceInstance,
) (provider.Advertiser, bool, error) {
	exhaustive := true

	for _, p := range r.Providers {
		a, ok, err := p.AdvertiserByDomain(ctx, res.Spec.Domain)
		if err != nil {
			exhaustive = false

			r.EventRecorder.Eventf(
				res,
				"Warning",
				"Error",
				"%s: %s",
				p.ID(),
				err.Error(),
			)

			if ctx.Err() != nil {
				return nil, false, ctx.Err()
			}
		}

		if !ok {
			continue
		}

		res.Status.ProviderID = p.ID()
		res.Status.ProviderDescription = p.Describe()
		res.Status.AdvertiserID = a.ID()

		if err := r.Client.Status().Update(ctx, res); err != nil {
			return nil, false, fmt.Errorf("unable to update resource status: %w", err)
		}

		r.EventRecorder.Eventf(
			res,
			"Normal",
			"ProviderAssociated",
			"the %q provider will be used to advertise this service instance",
			p.Describe(),
		)

		return a, true, nil
	}

	if exhaustive {
		var providers []string
		for _, p := range r.Providers {
			providers = append(providers, p.ID())
		}
		slices.Sort(providers)

		r.EventRecorder.Eventf(
			res,
			"Warning",
			"Ignored",
			"none of the configured providers can advertise on %q",
			res.Spec.Domain,
		)
	}

	return nil, false, nil
}

func (r *Reconciler) getAdvertiser(
	ctx context.Context,
	res *crd.DNSSDServiceInstance,
) (provider.Advertiser, bool, error) {
	for _, p := range r.Providers {
		if p.ID() != res.Status.ProviderID {
			continue
		}

		a, err := p.AdvertiserByID(ctx, res.Status.AdvertiserID)
		if err != nil {
			r.EventRecorder.Eventf(
				res,
				"Warning",
				"Error",
				"%s: %s",
				res.Status.ProviderID,
				err.Error(),
			)

			return nil, false, ctx.Err()
		}

		return a, true, nil
	}

	// This reconciler does not know about the provider that was assigned to the
	// resource. This is likely because the resource is managed by some other
	// instance of Proclaim.
	return nil, false, nil
}
