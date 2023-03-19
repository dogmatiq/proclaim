package reconciler

import (
	"context"

	"github.com/dogmatiq/proclaim/crd"
	"github.com/dogmatiq/proclaim/provider"
)

// getOrAssociateAdvertiser returns the advertiser used to
// advertise/unadvertise the given DNS-SD service instance.
func (r *Reconciler) getOrAssociateAdvertiser(
	ctx context.Context,
	res *crd.DNSSDServiceInstance,
) (provider.Advertiser, bool, error) {
	if res.Status.Provider != "" {
		return r.getAdvertiser(ctx, res)
	}
	return r.associateAdvertiser(ctx, res)
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
		a, ok, err := p.AdvertiserByDomain(ctx, res.Spec.Instance.Domain)
		if err != nil {
			crd.ProviderError(
				r.Manager,
				res,
				p.ID(),
				p.Describe(),
				err,
			)

			exhaustive = false

			if ctx.Err() != nil {
				return nil, false, ctx.Err()
			}
		}

		if !ok {
			continue
		}

		if err := r.update(
			res,
			crd.MergeCondition(crd.InstanceAdoptedCondition()),
			crd.UpdateProviderDescription(p.Describe()),
			crd.AssociateProvider(p.ID(), a.ID()),
		); err != nil {
			return nil, false, err
		}

		crd.InstanceAdopted(r.Manager, res)

		return a, true, nil
	}

	if exhaustive {
		crd.InstanceIgnored(r.Manager, res)

		if err := r.update(
			res,
			crd.MergeCondition(crd.InstanceIgnoredCondition()),
		); err != nil {
			return nil, false, err
		}
	}

	return nil, false, nil
}

func (r *Reconciler) getAdvertiser(
	ctx context.Context,
	res *crd.DNSSDServiceInstance,
) (provider.Advertiser, bool, error) {
	for _, p := range r.Providers {
		if p.ID() != res.Status.Provider {
			continue
		}

		// Make sure the provider's description is up-to-date.
		if err := r.update(
			res,
			crd.UpdateProviderDescription(p.Describe()),
		); err != nil {
			return nil, false, err
		}

		a, err := p.AdvertiserByID(ctx, res.Status.Advertiser)
		if err != nil {
			crd.ProviderError(
				r.Manager,
				res,
				p.ID(),
				p.Describe(),
				err,
			)
			return nil, false, ctx.Err()
		}

		return a, true, nil
	}

	// This reconciler does not know about the provider that is associated with
	// the resource. This is likely because the resource is managed by some
	// other instance of Proclaim.
	return nil, false, nil
}
