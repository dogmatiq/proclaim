package reconciler

import (
	"context"
	"fmt"

	"github.com/dogmatiq/proclaim/crd"
	"github.com/dogmatiq/proclaim/provider"
	"github.com/dogmatiq/proclaim/reconciler/internal/payload"
	"golang.org/x/exp/slices"
)

// getOrAssociateAdvertiser returns the advertiser used to
// advertise/unadvertise the given DNS-SD service instance.
func (r *Reconciler) getOrAssociateAdvertiser(
	ctx context.Context,
	res *crd.DNSSDServiceInstance,
) (provider.Advertiser, bool, error) {
	if len(res.Status.ProviderPayload) == 0 {
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

		res.Status.ProviderDescription = p.Describe()
		res.Status.ProviderPayload = payload.Marshal(p, a)
		res.Status.Status = crd.StatusAdvertising

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
	payload, err := payload.Unmarshal(res.Status.ProviderPayload)
	if err != nil {
		return nil, false, err
	}

	for _, p := range r.Providers {
		if p.ID() != payload.GetProviderId() {
			continue
		}

		// Make sure the provider's description is up-to-date.
		desc := p.Describe()
		if res.Status.ProviderDescription != desc {
			res.Status.ProviderDescription = desc

			if err := r.Client.Status().Update(ctx, res); err != nil {
				return nil, false, fmt.Errorf("unable to update resource status: %w", err)
			}
		}

		a, err := p.AdvertiserByID(ctx, payload.GetAdvertiserId())
		if err != nil {
			r.EventRecorder.Eventf(
				res,
				"Warning",
				"Error",
				"%s: %s",
				res.Status.ProviderDescription,
				err.Error(),
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
